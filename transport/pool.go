package transport

import (
	"net"
	"time"
)

var (
	DefaultConnectTimeout = time.Second * 5
	DefaultHandleTimeout  = time.Second * 5
)

type Option func(*ConnectionPool)

// 连接池
type Pooler interface {
	GetConn(addr string) (c net.Conn, err error)
}

type ConnectionPool struct {
	ConnectTimeout time.Duration // 连接超时设置
	HandleTimeout  time.Duration // handle 处理超时设置
}

func NewConnectionPool(opts ...Option) *ConnectionPool {
	c := &ConnectionPool{
		ConnectTimeout: DefaultConnectTimeout,
		HandleTimeout:  DefaultHandleTimeout,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func WithConnectionTimeout(t time.Duration) Option {
	return func(cp *ConnectionPool) {
		cp.ConnectTimeout = t
	}
}

func WithHandleTimeout(t time.Duration) Option {
	return func(cp *ConnectionPool) {
		cp.HandleTimeout = t
	}
}

// TODO: 暂时直接开新连接，之后需要池化
func (p *ConnectionPool) GetConn(addr string) (c net.Conn, err error) {
	return net.DialTimeout("tcp", addr, p.ConnectTimeout)
}
