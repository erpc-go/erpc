// Package workerpool 协程池
package workpool

import (
	"runtime"
	"sync"
	"time"
)

type Handler interface {
	Handle() error
}

type HandlerFunc func() error

func (h HandlerFunc) Handle() error {
	return h()
}

func New(maxWorkersCount int, maxIdleWorkerDuration time.Duration) *WorkerPool {
	return &WorkerPool{
		MaxWorkersCount:       maxWorkersCount,
		MaxIdleWorkerDuration: maxIdleWorkerDuration,
	}
}

// workerPool serves incoming connections via a pool of workers
// in FILO order, i.e. the most recently stopped worker will serve the next
// incoming connection.
//
// Such a scheme keeps CPU caches hot (in theory).
type WorkerPool struct {
	MaxWorkersCount int

	LogAllErrors bool

	MaxIdleWorkerDuration time.Duration

	lock         sync.Mutex
	workersCount int
	mustStop     bool

	ready []*workerChan

	stopCh chan struct{}

	workerChanPool sync.Pool
}

type workerChan struct {
	lastUseTime time.Time
	ch          chan Handler
}

func (wp *WorkerPool) Start() {
	if wp.stopCh != nil {
		panic("BUG: workerPool already started")
	}
	wp.stopCh = make(chan struct{})
	stopCh := wp.stopCh
	go func() {
		var scratch []*workerChan
		for {
			wp.clean(&scratch)
			select {
			case <-stopCh:
				return
			default:
				time.Sleep(wp.getMaxIdleWorkerDuration())
			}
		}
	}()
}

func (wp *WorkerPool) Stop() {
	if wp.stopCh == nil {
		panic("BUG: workerPool wasn't started")
	}
	close(wp.stopCh)
	wp.stopCh = nil

	// Stop all the workers waiting for incoming connections.
	// Do not wait for busy workers - they will stop after
	// serving the connection and noticing wp.mustStop = true.
	wp.lock.Lock()
	ready := wp.ready
	for i, ch := range ready {
		ch.ch <- nil
		ready[i] = nil
	}
	wp.ready = ready[:0]
	wp.mustStop = true
	wp.lock.Unlock()
}

func (wp *WorkerPool) getMaxIdleWorkerDuration() time.Duration {
	if wp.MaxIdleWorkerDuration <= 0 {
		return 10 * time.Second
	}
	return wp.MaxIdleWorkerDuration
}

func (wp *WorkerPool) clean(scratch *[]*workerChan) {
	maxIdleWorkerDuration := wp.getMaxIdleWorkerDuration()

	// Clean least recently used workers if they didn't serve connections
	// for more than maxIdleWorkerDuration.
	currentTime := time.Now()

	wp.lock.Lock()
	ready := wp.ready
	n := len(ready)
	i := 0
	for i < n && currentTime.Sub(ready[i].lastUseTime) > maxIdleWorkerDuration {
		i++
	}
	*scratch = append((*scratch)[:0], ready[:i]...)
	if i > 0 {
		m := copy(ready, ready[i:])
		for i = m; i < n; i++ {
			ready[i] = nil
		}
		wp.ready = ready[:m]
	}
	wp.lock.Unlock()

	// Notify obsolete workers to stop.
	// This notification must be outside the wp.lock, since ch.ch
	// may be blocking and may consume a lot of time if many workers
	// are located on non-local CPUs.
	tmp := *scratch
	for i, ch := range tmp {
		ch.ch <- nil
		tmp[i] = nil
	}
}

func (wp *WorkerPool) Serve(h Handler) bool {
	ch := wp.getCh()
	if ch == nil {
		return false
	}
	ch.ch <- h
	return true
}

var workerChanCap = func() int {
	// Use blocking workerChan if GOMAXPROCS=1.
	// This immediately switches Serve to WorkerFunc, which results
	// in higher performance (under go1.5 at least).
	if runtime.GOMAXPROCS(0) == 1 {
		return 0
	}

	// Use non-blocking workerChan if GOMAXPROCS>1,
	// since otherwise the Serve caller (Acceptor) may lag accepting
	// new connections if WorkerFunc is CPU-bound.
	return 1
}()

func (wp *WorkerPool) getCh() *workerChan {
	var ch *workerChan
	createWorker := false

	wp.lock.Lock()
	ready := wp.ready
	n := len(ready) - 1
	if n < 0 {
		if wp.workersCount < wp.MaxWorkersCount {
			createWorker = true
			wp.workersCount++
		}
	} else {
		ch = ready[n]
		ready[n] = nil
		wp.ready = ready[:n]
	}
	wp.lock.Unlock()

	if ch == nil {
		if !createWorker {
			return nil
		}
		vch := wp.workerChanPool.Get()
		if vch == nil {
			vch = &workerChan{
				ch: make(chan Handler, workerChanCap),
			}
		}
		ch = vch.(*workerChan)
		go func() {
			wp.workerFunc(ch)
			wp.workerChanPool.Put(vch)
		}()
	}
	return ch
}

func (wp *WorkerPool) release(ch *workerChan) bool {
	ch.lastUseTime = time.Now().Truncate(time.Second)
	wp.lock.Lock()
	if wp.mustStop {
		wp.lock.Unlock()
		return false
	}
	wp.ready = append(wp.ready, ch)
	wp.lock.Unlock()
	return true
}

func (wp *WorkerPool) workerFunc(ch *workerChan) {
	var h Handler

	var err error
	for h = range ch.ch {
		if h == nil {
			break
		}
		if err = h.Handle(); err != nil {
		}
		h = nil

		if !wp.release(ch) {
			break
		}
	}

	wp.lock.Lock()
	wp.workersCount--
	wp.lock.Unlock()
}
