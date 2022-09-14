package transport

import (
	"net"
	"time"
)

var (
	DefaultConnectTimeout = time.Second * 5
	DefaultHandleTimeout  = time.Second * 5
)

// 连接池接口
type Pooler interface {
	GetConn(addr string) (c net.Conn, err error)
}

// 具体连接池实现
type Option func(*ConnectionPool)

// TODO: 实现连接池
type ConnectionPool struct {
	ConnectTimeout time.Duration // 连接超时设置
	HandleTimeout  time.Duration // handle 处理超时设置

	conn map[string]net.Conn
}

func NewConnectionPool(opts ...Option) *ConnectionPool {
	c := &ConnectionPool{
		ConnectTimeout: DefaultConnectTimeout,
		HandleTimeout:  DefaultHandleTimeout,
		conn:           map[string]net.Conn{},
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
	conn, ok := p.conn[addr]
	if ok {
		return conn, nil
	}
	t, err := net.DialTimeout("tcp", addr, p.ConnectTimeout)
	p.conn[addr] = t
	return t, err
}
