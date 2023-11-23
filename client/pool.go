package client

import (
	"fmt"
	"sync"
	"time"
)

// Pool common connection pool
type Pool struct {
	// New create connection function
	New func() interface{}
	// Ping check connection is ok
	Ping func(interface{}) bool
	// Close close connection
	Close func(interface{})
	// Idle check connection idle timeout before Get
	Idle  time.Duration
	store chan *item
	mu    sync.Mutex
}

type item struct {
	data      interface{}
	heartbeat time.Time
}

// NewPool create a pool with capacity
func NewPool(initCap, maxCap int, newFunc func() interface{}) (*Pool, error) {
	if maxCap == 0 || initCap > maxCap {
		return nil, fmt.Errorf("invalid capacity settings")
	}
	p := new(Pool)
	p.store = make(chan *item, maxCap)
	if newFunc != nil {
		p.New = newFunc
	}
	for i := 0; i < initCap; i++ {
		v, err := p.create()
		if err != nil {
			return p, err
		}
		p.store <- &item{data: v, heartbeat: time.Now()}
	}
	return p, nil
}

// Len returns current connections in pool
func (p *Pool) Len() int {
	return len(p.store)
}

// RegisterChecker start a goroutine check status every interval time
func (p *Pool) RegisterChecker(interval time.Duration, check func(interface{}) bool) {
	if interval > 0 && check != nil {
		go func() {
			for {
				time.Sleep(interval)
				p.mu.Lock()
				if p.store == nil {
					// pool aleardy destroyed, exit
					p.mu.Unlock()
					return
				}
				l := p.Len()
				p.mu.Unlock()
				for idx := 0; idx < l; idx++ {
					select {
					case i := <-p.store:
						v := i.data
						if p.Idle > 0 && time.Now().Sub(i.heartbeat) > p.Idle {
							if p.Close != nil {
								p.Close(v)
							}
							continue
						}
						if !check(v) {
							if p.Close != nil {
								p.Close(v)
							}
							continue
						} else {
							select {
							case p.store <- i:
								continue
							default:
								if p.Close != nil {
									p.Close(v)
								}
							}
						}
					default:
						break
					}
				}
			}
		}()
	}
	return
}

// Get returns a conn form store or create one
func (p *Pool) Get() (interface{}, error) {
	if p.store == nil {
		// pool aleardy destroyed, returns new connection
		return p.create()
	}
	for {
		select {
		case i := <-p.store:
			v := i.data
			if p.Idle > 0 && time.Now().Sub(i.heartbeat) > p.Idle {
				if p.Close != nil {
					p.Close(v)
				}
				continue
			}
			if p.Ping != nil && p.Ping(v) == false {
				continue
			}
			return v, nil
		default:
			return p.create()
		}
	}
}

// Put set back conn into store again
func (p *Pool) Put(v interface{}) {
	if p.store == nil {
		// pool aleardy destroyed, returns empty
		return
	}
	select {
	case p.store <- &item{data: v, heartbeat: time.Now()}:
		return
	default:
		if p.Close != nil {
			p.Close(v)
		}
		return
	}
}

// Destroy clear all connections
func (p *Pool) Destroy() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.store == nil {
		// pool aleardy destroyed
		return
	}
	close(p.store)
	for i := range p.store {
		if p.Close != nil {
			p.Close(i.data)
		}
	}
	p.store = nil
}

func (p *Pool) create() (interface{}, error) {
	if p.New == nil {
		return nil, fmt.Errorf("Pool.New is nil, can not create connection")
	}
	return p.New(), nil
}
