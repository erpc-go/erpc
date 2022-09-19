package client

import (
	"errors"
	"net"

	"github.com/edte/erpc/transport"
)

// ConnFactory Implement
type connFactory struct {
	addr     string // server addr,such as: "127.0.0.1"
	protocol string // socket protocol type,such as "tcp"
}

func newConnFactory(protocol string, addr string) *connFactory {
	return &connFactory{
		addr:     addr,
		protocol: protocol,
	}
}

func (co *connFactory) Get() (c transport.Conn, err error) {
	return net.Dial(co.protocol, co.addr)
}

func (co *connFactory) Close(c transport.Conn) error {
	rawc, ok := c.(net.Conn)
	if !ok {
		return errors.New("not net conn interface")
	}
	return rawc.Close()
}

// ping 用 UDP 实现
func (co *connFactory) Ping(c transport.Conn) error {
	// client := NewClient()
	return nil
}
