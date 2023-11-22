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
	TCPRecvBufPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, defaultRecvBufSize)
		},
	}
}

const defaultRecvBufSize = 64 * 1024

// TCPRecvBufPool tcp接收缓冲队列
var TCPRecvBufPool *sync.Pool

// conn 支持Pipeline
// 每个conn对应三个goroutine
// 处理流程： client -> read_co -> work_co -> write_co -> client
// read_co: 读取客户端请求并放入cin
// work_co: 从cin读取请求处理并将结果写入cout
// write_co：从cout读取响应并发送给客户端
type conn struct {
	server     *TCPServer
	rwc        net.Conn
	checker    Checker
	cancelCtx  context.CancelFunc
	remoteAddr string
	mu         sync.Mutex
	cin        chan []byte
	cout       chan []byte
	wg         sync.WaitGroup
}

// CtxKey ctx内部存放客户端地址的key
type CtxKey string

// ClientAddr ctx内部存放客户端地址的key
const (
	ClientAddr CtxKey = "clientAddr"
	ServerAddr CtxKey = "serverAddr"
)

func (c *conn) readRequests(ctx context.Context) {
	log.Raw("tcp read goroutine start")
	defer func() {
		log.Raw("tcp read goroutine return")
		c.cancelCtx()
		c.wg.Done()
	}()

	buffer := TCPRecvBufPool.Get().([]byte)
	defer func() {
		TCPRecvBufPool.Put(buffer)
	}()

	var nRead int
	for !c.server.closing {
		c.rwc.SetDeadline(time.Now().Add(c.server.idleTimeout))
		n, err := c.rwc.Read(buffer[nRead:])
		atomic.AddUint64(&RecvBytes, uint64(n))
		atomic.AddUint64(&RecvPkgs, 1)
		if err != nil {
			log.Raw("tcp read fail: ", n, err)
			if c.server.closing {
				break
			}
			log.Raw("tcp read routine canceled context")
			return
		}
		nRead += n

		if nRead >= cap(buffer) {
			log.Raw("tcp read too big, read %d bytes, expand twice", nRead)
			tmpBuffer := make([]byte, nRead*2)
			copy(tmpBuffer, buffer[:nRead])
			buffer = tmpBuffer
		}

		var readIndex int
		for {
			pkgLen, err := c.server.checker.Check(buffer[readIndex:nRead])
			if err != nil || pkgLen < 0 {
				log.Raw("tcp check fail: ", pkgLen, err)
				return
			}

			if readIndex+pkgLen > nRead {
				log.Raw("tcp check fail, pkglen:%d > nRead:%d", readIndex+pkgLen, nRead)
				return
			}
			if !c.server.limiter.Allow() {
				return
			}

			if pkgLen == 0 {
				// 未收完，分包
				log.Raw("tcp uncomplete packet: ", readIndex, nRead)
				break
			}

			// 接收完成
			req := make([]byte, pkgLen)
			copy(req, buffer[readIndex:readIndex+pkgLen])
			select {
			case <-ctx.Done():
				log.Raw("tcp read routine context done:", ctx.Err())
				return
			case c.cin <- req:
				readIndex += pkgLen
				log.Raw("read %v bytes from %v", pkgLen, c.remoteAddr)
			}

			if readIndex >= nRead {
				// 正常包
				break
			}

			// 多收，粘包, 继续check
			log.Raw("tcp stick packet: ", readIndex, nRead)
		}

		if readIndex > 0 {
			if readIndex < nRead {
				copy(buffer, buffer[readIndex:nRead])
			}
			nRead -= readIndex
		}
	}

	if c.server.closing {
		log.Raw("tcp connection closing, sleep 5 seconds")
		time.Sleep(time.Second * 5)
	}
}

func (c *conn) handle(ctx context.Context) {
	log.Raw("tcp handle goroutine start")
	defer func() {
		log.Raw("tcp handle goroutine return")
		c.cancelCtx()
		c.wg.Done()
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
			log.Raw("tcp handle routine context done:", ctx.Err())
			return
		case req := <-c.cin:
			go func() {
				log.Raw("tcp handle business goroutine start")
				defer func() {
					log.Raw("tcp handle business goroutine return")
				}()
				subCtx, cancel := context.WithTimeout(ctx, c.server.msgTimeout)
				rsp, err := c.server.handler.Serve(subCtx, req)
				cancel()
				if err != nil {
					log.Raw(err.Error())
					return
				}
				select {
				case <-ctx.Done():
					log.Raw("tcp handle business goroutine context done")
					return
				case c.cout <- rsp:
					log.Raw("tcp handle business goroutine write rsp to cout channel")
				}
			}()
		}
	}
}

func (c *conn) writeResponses(ctx context.Context) {
	log.Raw("tcp write goroutine start")
	defer func() {
		log.Raw("tcp write goroutine return")
		c.cancelCtx()
		c.wg.Done()
	}()

	for {
		select {
		case <-ctx.Done():
			log.Raw("tcp write routine context done:", ctx.Err())
			return
		case rsp := <-c.cout:
			n, err := c.rwc.Write(rsp)
			if err != nil {
				log.Raw(err.Error())
				return
			}
			log.Raw("write %v bytes to %v", n, c.remoteAddr)
			atomic.AddUint64(&SendBytes, uint64(n))
			atomic.AddUint64(&SendPkgs, 1)
		}
	}
}

func (c *conn) serve(ctx context.Context) {
	c.remoteAddr = c.rwc.RemoteAddr().String()
	ctx = context.WithValue(ctx, ClientAddr, c.remoteAddr)
	ctx, cancelCtx := context.WithCancel(ctx)
	c.cancelCtx = cancelCtx
	log.Raw("accept tcp connection from:", c.remoteAddr)
	defer c.rwc.Close()

	c.wg.Add(3)
	go c.readRequests(ctx)
	go c.handle(ctx)
	go c.writeResponses(ctx)
	c.wg.Wait()
	log.Raw("tcp connection destroyed")
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

// Accept tcp listener accept function
func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

// TCPServer tcp服务结构体
type TCPServer struct {
	addr        *net.TCPAddr // TCP address to listen on, ":http" if empty
	handler     Handler      // handler to invoke, http.DefaultServeMux if nil
	checker     Checker
	limiter     Limiter
	conn        tcpKeepAliveListener
	msgTimeout  time.Duration // 消息处理的最大时长
	idleTimeout time.Duration // 长链接空闲时间
	closing     bool
	stop        chan int
}

// ListenAndServe 启动tcp服务
func (srv *TCPServer) ListenAndServe() error {
	ln, err := net.ListenTCP("tcp", srv.addr)
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	return srv.Serve(tcpKeepAliveListener{ln})
}

// Serve 处理请求
func (srv *TCPServer) Serve(l net.Listener) error {
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
				log.Raw("tcp: Accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0
		c := srv.newConn(rw)
		go c.serve(ctx)
	}

	log.Raw("tcp server closing")
	<-srv.stop
	log.Raw("tcp server stopped")
	time.Sleep(time.Second)
	return nil
}

func (srv *TCPServer) newConn(rwc net.Conn) *conn {
	c := &conn{
		server: srv,
		rwc:    rwc,
		cin:    make(chan []byte, 10),
		cout:   make(chan []byte, 11),
	}
	return c
}

// Fork 热重启fork子进程，老进程停止接收请求
func (srv *TCPServer) Fork() (int, error) {
	log.Raw("fork tcp server")
	if srv.closing {
		return 0, fmt.Errorf("tcp server already forked")
	}
	file, err := srv.conn.File()
	if err != nil {
		return 0, fmt.Errorf("tcp server get conn file fail:%s", err)
	}
	// ! remember to release it
	defer file.Close()
	return StartNewProcess(file.Fd(), 0, 0)
}

// Shutdown 热重启，让老进程退出
func (srv *TCPServer) Shutdown() error {
	log.Raw("shutdown tcp server, wait for 10 seconds")
	if !srv.closing {
		srv.closing = true
		srv.idleTimeout = 0
		srv.conn.SetDeadline(time.Now())
		log.Raw("tcp listener set deadline")
	}
	time.Sleep(time.Second * 10)
	srv.stop <- 1

	return nil
}
