package net

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/erpc-go/erpc/server/workpool"
	"github.com/erpc-go/log"
)

// udp最大接收size
const (
	MaxUDPPkg = 65536
)

// UDPRequest 单个udp请求的数据结构
type UDPRequest struct {
	req                 []byte
	localAddr, peerAddr *net.UDPAddr
	handler             Handler
	msgTimeout          time.Duration
	conn                *net.UDPConn
}

// Handle 处理udp请求
func (r *UDPRequest) Handle() error {
	ctx, cancel := context.WithTimeout(context.Background(), r.msgTimeout)
	defer cancel()
	ctx = context.WithValue(ctx, ClientAddr, r.peerAddr.String())
	defer func() {
		if e := recover(); e != nil {
			buf := make([]byte, RecoverStackSize)
			buf = buf[:runtime.Stack(buf, false)]
			log.Raw("%v: %v\n%s", *r.peerAddr, e, buf)
		}
	}()
	rsp, e := r.handler.Serve(ctx, r.req)
	if e == nil && len(rsp) > 0 {
		if len(rsp) >= MaxUDPPkg {
		}
		if n, err := r.conn.WriteToUDP(rsp, r.peerAddr); err == nil {
			log.Raw("write %v bytes\n", n)
			atomic.AddUint64(&SendBytes, uint64(n))
			atomic.AddUint64(&SendPkgs, 1)
		} else {
			log.Raw(e.Error())
		}
	} else {
		log.Raw("rsp err %v, rsp len %v", e, len(rsp))
	}
	return nil
}

// UDPServer defines parameters for running an UDP server.
type UDPServer struct {
	addr       *net.UDPAddr // UDP address to listen on
	handler    Handler      // handler to invoke
	checker    Checker
	limiter    Limiter
	conn       *net.UDPConn
	msgTimeout time.Duration // 消息处理的最大时长
	workerpool *workpool.WorkerPool
	closing    bool
	stop       chan int
}

// ListenAndServe listens on the UDP network address srv.Addr and then
// calls Serve to handle requests on incoming connections.  If
// srv.Addr is blank, ":http" is used.
func (srv *UDPServer) ListenAndServe() error {
	conn, e := net.ListenUDP("udp", srv.addr)
	if e != nil {
		panic(e)
	}
	srv.conn = conn
	defer srv.conn.Close()

	return srv.Serve()
}

// Serve accepts incoming connections on the Listener l, creating a
// new service thread for each.  The service threads read requests and
// then call srv.Handler to reply to them.
func (srv *UDPServer) Serve() error {
	var tempDelay time.Duration // how long to sleep on accept failure
	recvBuf := make([]byte, MaxUDPPkg)
	for !srv.closing {
		n, raddr, e := srv.conn.ReadFromUDP(recvBuf)
		atomic.AddUint64(&RecvBytes, uint64(n))
		atomic.AddUint64(&RecvPkgs, 1)
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
				log.Raw("udp ReadFromUDP error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return e
		}

		log.Raw("read %v bytes\n", n)
		if !srv.limiter.Allow() {
			continue
		}
		if num, e := srv.checker.Check(recvBuf[:n]); num == 0 || e != nil {
			log.Raw("bad pkg, recv %d, error %v, num %d", n, e, num)
			continue
		}
		request := &UDPRequest{
			req:        make([]byte, n, n),
			localAddr:  srv.addr,
			peerAddr:   raddr,
			handler:    srv.handler,
			msgTimeout: srv.msgTimeout,
			conn:       srv.conn,
		}
		copy(request.req, recvBuf[:n])
		if !srv.workerpool.Serve(request) {
			log.Raw("workerpool over ratelimit")
			srv.conn.WriteToUDP([]byte("over ratelimit"), raddr)
		}
	}

	log.Raw("udp server closing")
	<-srv.stop
	log.Raw("udp server stopped")
	time.Sleep(time.Second)
	return nil
}

// Fork 热重启fork子进程，老进程停止接收请求
func (srv *UDPServer) Fork() (int, error) {
	log.Raw("fork udp server")
	if srv.closing {
		return 0, fmt.Errorf("udp server already forked")
	}

	file, err := srv.conn.File()
	if err != nil {
		return 0, fmt.Errorf("udp server get conn file fail:%s", err)
	}
	// ! remember to release it
	defer file.Close()
	return StartNewProcess(0, file.Fd(), 0)
}

// Shutdown 热重启，让老进程退出
func (srv *UDPServer) Shutdown() error {
	log.Raw("shutdown udp server, wait for 10 seconds")
	if !srv.closing {
		srv.closing = true
		srv.conn.SetReadDeadline(time.Now())
	}
	time.Sleep(time.Second * 10)
	srv.stop <- 1
	return nil
}
