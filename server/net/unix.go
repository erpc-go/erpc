package net

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/erpc-go/log"
)

func init() {
	UnixRecvBufPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, defaultRecvBufSize)
		},
	}
}

var UnixRecvBufPool *sync.Pool

// conn 支持Pipeline
// 每个conn对应三个goroutine
// 处理流程： client -> read_co -> work_co -> write_co -> client
// read_co: 读取客户端请求并放入cin
// work_co: 从cin读取请求处理并将结果写入cout
// write_co：从cout读取响应并发送给客户端
type unixconn struct {
	server     *UnixServer
	rwc        net.Conn
	checker    Checker
	cancelCtx  context.CancelFunc
	remoteAddr string
	mu         sync.Mutex
	cin        chan []byte
	cout       chan []byte
}

func (c *unixconn) serve(ctx context.Context) {
	c.remoteAddr = c.rwc.RemoteAddr().String()
	ctx = context.WithValue(ctx, ClientAddr, c.remoteAddr)
	ctx, cancelCtx := context.WithCancel(ctx)
	c.cancelCtx = cancelCtx
	log.Raw("accept unix connection from:", c.remoteAddr)
	defer c.rwc.Close()

	var wg sync.WaitGroup
	wg.Add(3)
	{
		// 读取客户端请求
		go func(ctx context.Context, ch chan<- []byte) {
			defer wg.Done()
			defer c.cancelCtx()

			buffer := UnixRecvBufPool.Get().([]byte)
			defer func() {
				UnixRecvBufPool.Put(buffer)
			}()

			var nRead int
			for !c.server.closing {
				c.rwc.SetDeadline(time.Now().Add(c.server.idleTimeout))
				n, e := c.rwc.Read(buffer[nRead:])
				atomic.AddUint64(&RecvBytes, uint64(n))
				atomic.AddUint64(&RecvPkgs, 1)
				if e != nil || n == 0 {
					log.Raw("unix read fail: ", n, e)
					if c.server.closing {
						break
					}
					log.Raw("unix read routine canceled context")
					return
				}
				nRead += n
				if nRead >= cap(buffer) {
					log.Raw("unix read too big, read %d bytes, expand twice", nRead)
					tmpBuffer := make([]byte, nRead*2)
					copy(tmpBuffer, buffer[:nRead])
					buffer = tmpBuffer
				}

				var readIndex int
				for {
					pkgLen, e := c.server.checker.Check(buffer[readIndex:nRead])
					if e != nil || pkgLen < 0 {
						log.Raw("unix check fail: ", pkgLen, e)
						return
					}

					if pkgLen > 0 {
						// 接收完成
						if readIndex+pkgLen > nRead {
							log.Raw("unix check fail, pkglen:%d > nRead:%d", readIndex+pkgLen, nRead)
							return
						}
						if !c.server.limiter.Allow() {
							return
						}

						req := make([]byte, pkgLen)
						copy(req, buffer[readIndex:readIndex+pkgLen])
						select {
						case <-ctx.Done():
							log.Raw("unix read routine context done:", ctx.Err())
							return
						case ch <- req:
							readIndex += pkgLen
							log.Raw("read %v bytes from %v", pkgLen, c.remoteAddr)
						}

						if readIndex < nRead {
							// 多收，粘包, 继续check
							log.Raw("unix stick packet: ", readIndex, nRead)
						} else {
							// 正常包
							break
						}
					} else {
						// 未收完，分包
						log.Raw("unix uncomplete packet: ", readIndex, nRead)
						break
					}
				}
				if readIndex > 0 {
					if readIndex < nRead {
						copy(buffer, buffer[readIndex:nRead])
					}
					nRead -= readIndex
				}
			}
			if c.server.closing {
				log.Raw("unix connection closing, sleep 5 seconds")
				time.Sleep(time.Second * 5)
			}
		}(ctx, c.cin)
	}
	{
		// 处理客户端请求
		go func(ctx context.Context, in <-chan []byte, out chan<- []byte) {
			defer func() {
				c.cancelCtx()
				wg.Done()
				if err := recover(); err != nil {
					c.rwc.SetDeadline(time.Now())
					buf := make([]byte, RecoverStackSize)
					buf = buf[:runtime.Stack(buf, false)]
					log.Raw("%v: %v\n%s", c.remoteAddr, err, buf)
				}
			}()

			for {
				select {
				case <-ctx.Done():
					log.Raw("unix handle routine context done:", ctx.Err())
					return
				case req := <-in:
					go func() {
						subctx, cancel := context.WithTimeout(ctx, c.server.msgTimeout)
						rsp, e := c.server.handler.Serve(subctx, req)
						cancel()
						if e != nil {
							log.Raw(e.Error())
							return
						}
						c.cout <- rsp
					}()
				}
			}
		}(ctx, c.cin, c.cout)
	}
	{
		// 向客户端回包
		go func(ctx context.Context, ch <-chan []byte) {
			defer wg.Done()
			defer c.cancelCtx()
			for {
				select {
				case <-ctx.Done():
					log.Raw("unix write routine context done:", ctx.Err())
					return
				case rsp := <-ch:
					n, e := c.rwc.Write(rsp)
					if e != nil {
						log.Raw(e.Error())
						return
					}
					log.Raw("write %v bytes to %v", n, c.remoteAddr)
					atomic.AddUint64(&SendBytes, uint64(n))
					atomic.AddUint64(&SendPkgs, 1)
				}
			}
		}(ctx, c.cout)
	}
	wg.Wait()
	log.Raw("unix connection destroyed")
}

type UnixServer struct {
	addr        *net.UnixAddr // Unix address to listen on, ":http" if empty
	handler     Handler       // handler to invoke, http.DefaultServeMux if nil
	checker     Checker
	limiter     Limiter
	conn        *net.UnixListener
	msgTimeout  time.Duration // 消息处理的最大时长
	idleTimeout time.Duration // 长链接空闲时间
	closing     bool
	stop        chan int
}

func (srv *UnixServer) ListenAndServe() error {
	ln, err := net.ListenUnix("unix", srv.addr)
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	return srv.Serve(ln)
}

// Serve 处理请求
func (srv *UnixServer) Serve(l net.Listener) error {
	ctx := context.WithValue(context.Background(), ServerAddr, srv.addr.String())
	var tempDelay time.Duration // how long to sleep on accept failure
	for !srv.closing {
		rw, e := l.Accept()
		if e != nil {
			log.Raw(e.Error())
			if srv.closing {
				break
			}
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Raw("unix: Accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0
		c := srv.newConn(rw)
		go c.serve(ctx)
	}

	log.Raw("unix server closing")
	<-srv.stop
	log.Raw("unix server stopped")
	time.Sleep(time.Second)
	return nil
}

func (srv *UnixServer) newConn(rwc net.Conn) *unixconn {
	c := &unixconn{
		server: srv,
		rwc:    rwc,
		cin:    make(chan []byte, 10),
		cout:   make(chan []byte, 11),
	}
	return c
}

// Fork 热重启fork子进程，老进程停止接收请求
func (srv *UnixServer) Fork() (int, error) {
	log.Raw("fork unix server")
	if srv.closing {
		return 0, fmt.Errorf("unix server already forked")
	}
	file, err := srv.conn.File()
	if err != nil {
		return 0, fmt.Errorf("unix server get conn file fail:%s", err)
	}
	// ! remember to release it
	defer file.Close()
	return StartNewProcess(0, 0, file.Fd())
}

// Shutdown 热重启，让老进程退出
func (srv *UnixServer) Shutdown() error {
	log.Raw("shutdown unix server, wait for 10 seconds")
	if !srv.closing {
		srv.closing = true
		srv.idleTimeout = 0
		srv.conn.SetDeadline(time.Now())
		log.Raw("unix listener set deadline")
	}
	time.Sleep(time.Second * 10)
	srv.stop <- 1

	return nil
}
