package transport

import (
	"context"
	"net"
)

// 连接
type Conn interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
}

// 连接工厂
type ConnFactory interface {
	Get(ctx context.Context) (c Conn, err error)
	Close(c Conn) error
	Ping(c Conn) error
}

// 连接池接口
type Pooler interface {
	// get conn
	Get(ctx context.Context) (c Conn, err error)
	// put conn
	Put(c Conn) (err error)
	// close conn
	Close(c Conn) (err error)
	// close all conn
	Release() (err error)
	// get valid connn number
	Len() int
}

// 具体连接池实现
type Option func(*ConnectionPool)

// TODO: 实现连接池
type ConnectionPool struct {
	factory ConnFactory
}

func NewConnectionPool(opts ...Option) *ConnectionPool {
	c := &ConnectionPool{}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// get conn
func (co *ConnectionPool) Get(ctx context.Context) (c Conn, err error) {
	panic("not implemented") // TODO: Implement
}

// put conn
func (co *ConnectionPool) Put(c Conn) (err error) {
	panic("not implemented") // TODO: Implement
}

// close conn
func (co *ConnectionPool) Close(c Conn) (err error) {
	panic("not implemented") // TODO: Implement
}

// close all conn
func (co *ConnectionPool) Release() (err error) {
	panic("not implemented") // TODO: Implement
}

// get valid connn number
func (co *ConnectionPool) Len() int {
	panic("not implemented") // TODO: Implement
}
