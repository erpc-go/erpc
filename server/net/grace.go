package net

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/erpc-go/log"
)

// graceful const
const (
	GracefulEnvironKey     = "IsGracefulCat"
	GracefulEnvironStr     = GracefulEnvironKey + "=1"
	GracefulTCPListenerFd  = 3 // 0 stdin 1 stdout 2 stderr
	GracefulUDPListenerFd  = 4
	GracefulUnixListenerFd = 5
)

// GracefulRestart 支持热重启服务需实现的接口
type GracefulRestart interface {
	Fork() (int, error)
	Shutdown() error
}

// HandleSignals 监听信号
func HandleSignals(srv GracefulRestart) {
	signalChan := make(chan os.Signal)

	signal.Ignore(syscall.SIGPIPE) // 忽略 SIGPIPE 信号，避免对端 crash 导致本服务异常退出
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGUSR2, syscall.SIGSEGV, syscall.SIGINT)
	log.Raw("server notify signal: SIGTERM SIGUSR2 SIGSEGV SIGINT")

	for {
		sig := <-signalChan
		switch sig {
		case syscall.SIGTERM:
			log.Raw("receive SIGTERM signal, shutdown server")
			srv.Shutdown()
		case syscall.SIGSEGV:
			log.Raw("receive SIGSEGV signal, shutdown server")
		case syscall.SIGUSR2:
			log.Raw("receive SIGUSR2 signal, graceful restarting server")
			if pid, err := srv.Fork(); err != nil {
				log.Raw("start new process failed: %v, continue serving", err)
			} else {
				log.Raw("start new process succeed, the new pid is %d", pid)
				// 不再主动退出了，由子进程来触发TERM
				// srv.Shutdown()
			}
		case syscall.SIGINT:
			log.Raw("receive SIGINT signal, shutdown server")
			srv.Shutdown()
		default:
		}
	}
}

// StartNewProcess fork子进程，传入listener复用fd
func StartNewProcess(tcpfd, udpfd, unixfd uintptr) (int, error) {
	log.Raw("graceful start new process, tcp fd:%v, udp fd:%v, unix fd:%v",
		tcpfd, udpfd, unixfd)
	envs := []string{}
	for _, value := range os.Environ() {
		if value != GracefulEnvironStr {
			envs = append(envs, value)
		}
	}
	envs = append(envs, GracefulEnvironStr)

	execSpec := &syscall.ProcAttr{
		Env:   envs,
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), tcpfd, udpfd, unixfd},
	}

	fork, err := syscall.ForkExec(os.Args[0], os.Args, execSpec)
	if err != nil {
		return 0, fmt.Errorf("failed to forkexec: %v", err)
	}

	return fork, nil
}

// GetTCPListener 获取tcp listener
func GetTCPListener(addr *net.TCPAddr) (*net.TCPListener, error) {
	var ln *net.TCPListener
	var err error

	if os.Getenv(GracefulEnvironKey) != "" {
		log.Raw("tcp server get listener from os file")
		file := os.NewFile(GracefulTCPListenerFd, "")
		var listener net.Listener
		listener, err = net.FileListener(file)
		if err != nil {
			err = fmt.Errorf("net.FileListener error: %v", err)
			return nil, err
		}
		var ok bool
		if ln, ok = listener.(*net.TCPListener); !ok {
			return nil, fmt.Errorf("net.FileListener is not TCPListener")
		}
	} else {
		log.Raw("tcp server create listener from tcpaddr")
		ln, err = net.ListenTCP("tcp", addr)
		if err != nil {
			err = fmt.Errorf("net.ListenTCP error: %v", err)
			return nil, err
		}
	}
	return ln, nil
}

// GetUDPListener 获取udp listener
func GetUDPListener(addr *net.UDPAddr) (*net.UDPConn, error) {
	var ln *net.UDPConn
	var err error

	if os.Getenv(GracefulEnvironKey) != "" {
		log.Raw("udp server get listener from os file")
		file := os.NewFile(GracefulUDPListenerFd, "")
		var listener net.Conn
		listener, err = net.FileConn(file)
		if err != nil {
			err = fmt.Errorf("net.FileConn error: %v", err)
			return nil, err
		}
		var ok bool
		if ln, ok = listener.(*net.UDPConn); !ok {
			return nil, fmt.Errorf("net.FileConn is not UDPConn")
		}
	} else {
		log.Raw("udp server create listener from udpaddr")
		ln, err = net.ListenUDP("udp", addr)
		if err != nil {
			err = fmt.Errorf("net.ListenUDP error: %v", err)
			return nil, err
		}
	}
	return ln, nil
}

func GetUnixListener(addr *net.UnixAddr) (*net.UnixListener, error) {
	var ln *net.UnixListener
	var err error

	if os.Getenv(GracefulEnvironKey) != "" {
		log.Raw("unix server get listener from os file")
		file := os.NewFile(GracefulUnixListenerFd, "")
		var listener net.Listener
		listener, err = net.FileListener(file)
		if err != nil {
			err = fmt.Errorf("net.FileConn error: %v", err)
			return nil, err
		}
		var ok bool
		if ln, ok = listener.(*net.UnixListener); !ok {
			return nil, fmt.Errorf("net.FileConn is not UnixListener")
		}
	} else {
		log.Raw("unix server create listener from unixaddr")
		if err := os.Remove(addr.Name); err != nil && !os.IsNotExist(err) {
			panic(fmt.Sprintf("remove unix socket file %q: %s", addr.Name, err))
		}
		ln, err = net.ListenUnix("unix", addr)
		if err != nil {
			err = fmt.Errorf("net.ListenUnix error: %v", err)
			return nil, err
		}
	}
	return ln, nil
}
