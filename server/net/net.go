package net

import (
	"context"
	"net"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/erpc-go/erpc/server/workpool"
	"github.com/erpc-go/log"
	limit "github.com/erpc-go/ratelimit"
)

// 父进程id
var ppid = os.Getppid()

// 消息处理的最大超时时间
var defaultMsgTimeout = 800 * time.Millisecond

// 是否开启debug模式
var defaultEnableDebugmode = false

var defaultWorkerPool *workpool.WorkerPool

var defaultEnableGracefulRestart bool

var defaultIdleTimeout time.Duration

// RecoverStackSize panic时捕获的stack大小，16M
const RecoverStackSize = 16 * 1 << 10

// Server server interface with ListenAndServe function
type Server interface {
	ListenAndServe() error
}

// Handler server请求处理器
type Handler interface {
	Serve(context.Context, []byte) ([]byte, error)
}

// Limiter
type Limiter interface {
	Wait()       // 同步睡眠
	Allow() bool // 异步返回 bool
}

// HandlerFunc 适配器，将原生的函数适配为Handler
// 加入f为HandlerFunc类型的函数，则HandlerFunc(f)为一个
// 内部自动调用f的Handler对象
type HandlerFunc func(context.Context, []byte) ([]byte, error)

// Serve 由handler func转换成 handler interface
func (h HandlerFunc) Serve(ctx context.Context, req []byte) ([]byte, error) {
	return h(ctx, req)
}

// Checker 是包完整性检查接口
// 返回值：
//
//	0, nil: 包未接收完
//	>0, nil: 包以接收完并返回对应的包长度
//	0, err: 包错误
type Checker interface {
	Check([]byte) (int, error)
}

// CheckerFunc function to interface
type CheckerFunc func([]byte) (int, error)

// Check Checker interface
func (c CheckerFunc) Check(data []byte) (int, error) {
	return c(data)
}

// DefaultChecker 默认校验函数，直接返回包大小，只适合于udp
var DefaultChecker Checker = CheckerFunc(func(data []byte) (int, error) {
	return len(data), nil
})

// HandlerWrapper 包装器，用于统一attr属性上报
type HandlerWrapper struct {
	h Handler
}

// Serve HandlerWrapper handler
func (w *HandlerWrapper) Serve(ctx context.Context, req []byte) ([]byte, error) {
	rsp, err := w.h.Serve(ctx, req)
	return rsp, err
}

// CheckerWrapper 包装器，用于统一attr属性上报
type CheckerWrapper struct {
	c Checker
}

// Check CheckerWrapper handler
func (w *CheckerWrapper) Check(data []byte) (int, error) {
	n, err := w.c.Check(data)
	return n, err
}

func killParentProcess() {
	// 如果是热重启，通知父进程退出
	if ppid != 1 && os.Getenv(GracefulEnvironKey) != "" {
		if err := syscall.Kill(ppid, syscall.SIGINT); err != nil {
			log.Raw("failed to close parent: %+v\n", err)
		}
	}
}

// ListenAndServeUDP 监听UDP地址addr，内部调用checker进行包完整性检查和拆包
// 调用handler处理请求，msgTimeout为消息的超时时间
func ListenAndServeUDP(addr string, checker Checker, handler Handler, limiter limit.Limiter) {
	listenAddr, e := net.ResolveUDPAddr("udp4", addr)
	log.Raw("udp addr: %s\n", addr)
	if e != nil {
		panic("invalid listen addr")
	}
	server := &UDPServer{
		addr:       listenAddr,
		handler:    handler,
		checker:    checker,
		limiter:    limiter,
		msgTimeout: defaultMsgTimeout,
		workerpool: defaultWorkerPool,
		closing:    false,
		stop:       make(chan int),
	}
	if defaultEnableGracefulRestart {
		log.Raw("udp server EnableGracefulRestart\n")
		go HandleSignals(server)
		ln, e := GetUDPListener(listenAddr)
		if e != nil {
			panic(e)
		}
		server.conn = ln

		go killParentProcess()

		server.Serve()
	} else {
		server.ListenAndServe()
	}
}

// ListenAndServeTCP 监听TCP地址addr，内部调用checker进行包完整性检查和拆包
// 调用handler处理请求，msgTimeout为消息的超时时间
func ListenAndServeTCP(addr string, checker Checker, handler Handler, limiter Limiter) {
	listenAddr, e := net.ResolveTCPAddr("tcp4", addr)
	log.Raw("tcp addr: %s\n", listenAddr)
	if e != nil {
		panic("invalid listen addr")
	}
	server := &TCPServer{
		addr:        listenAddr,
		handler:     handler,
		checker:     checker,
		msgTimeout:  defaultMsgTimeout,
		idleTimeout: defaultIdleTimeout,
		limiter:     limiter,
		closing:     false,
		stop:        make(chan int),
	}
	if defaultEnableGracefulRestart {
		log.Raw("tcp server EnableGracefulRestart\n")
		go HandleSignals(server)
		ln, e := GetTCPListener(listenAddr)
		if e != nil {
			panic(e)
		}

		go killParentProcess()

		server.conn = tcpKeepAliveListener{ln}
		server.Serve(server.conn)
	} else {
		server.ListenAndServe()
	}
}

func ListenAndServeUnix(addr string, checker Checker, handler Handler, limiter Limiter, IdleTimeout time.Duration) {
	listenAddr, e := net.ResolveUnixAddr("unix", addr)
	log.Raw("unix socket addr: %s\n", addr)
	if e != nil {
		panic("invalid listen addr")
	}
	server := &UnixServer{
		addr:        listenAddr,
		handler:     handler,
		checker:     checker,
		limiter:     limiter,
		msgTimeout:  defaultMsgTimeout,
		idleTimeout: IdleTimeout,
	}
	if defaultEnableGracefulRestart {
		log.Raw("unix server EnableGracefulRestart\n")
		go HandleSignals(server)
		ln, e := GetUnixListener(listenAddr)
		if e != nil {
			panic(e)
		}

		go killParentProcess()

		server.conn = ln
		server.Serve(server.conn)
	} else {
		server.ListenAndServe()
	}
}

// ListenAndServe 同时监听TCP和UDP地址addr， 内部调用ListenAndServeUDP，ListenAndServeTCP
func ListenAndServe(addr string, checker Checker, handler Handler, limiter limit.Limiter) {
	if defaultEnableGracefulRestart {
		defaultEnableGracefulRestart = false // 同时监听tcp udp，不支持热重启
		log.Raw("tcp/udp all server do not support EnableGracefulRestart\n")
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		ListenAndServeUDP(addr, checker, handler, limiter)
		wg.Done()
	}()
	go func() {
		ListenAndServeTCP(addr, checker, handler, limiter)
		wg.Done()
	}()
	wg.Wait()
}

// Init 服务初始化，设置msg总超时，开启调试模式，日志等级，创建协程池
func Init(msgTimeout time.Duration, IdleTimeout time.Duration, enableDebugMode bool, logLevel uint8, MaxWorkerCount int, EnableGracefulRestart bool) {
	defaultMsgTimeout = msgTimeout
	defaultEnableDebugmode = enableDebugMode
	defaultWorkerPool = workpool.New(MaxWorkerCount, time.Minute)
	defaultEnableGracefulRestart = enableDebugMode
	defaultIdleTimeout = IdleTimeout
}
