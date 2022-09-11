package transport

import (
	"net"
	"time"
)

// 连接池
type Pooler interface {
	GetConn(addr string) (c net.Conn, err error)
}

type ConnectionPool struct {
	ConnectTimeout time.Duration // 连接超时设置
}

func NewConnectionPool(t time.Duration) *ConnectionPool {
	return &ConnectionPool{
		ConnectTimeout: t,
	}
}

// TODO:
// 暂时直接开新连接，之后需要池化
func (p *ConnectionPool) GetConn(addr string) (c net.Conn, err error) {
	return net.Dial("tcp", addr)
}
