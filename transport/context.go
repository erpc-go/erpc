package transport

import (
	"net"

	"github.com/edte/erpc/protocol"
)

// 连接上下文
type Context struct {
	RequestConn  *protocol.Request
	ResponseConn *protocol.Response

	Request  interface{}
	Response interface{}

	conn *net.Conn
}

func NewContext(c *net.Conn) *Context {
	return &Context{
		RequestConn:  &protocol.Request{},
		ResponseConn: &protocol.Response{},
		conn:         c,
	}
}

func (c *Context) ReadRequest() (err error) {
	return c.RequestConn.DecodeFrom(*c.conn)
}

func (c *Context) SendRequest() (err error) {
	return c.RequestConn.EncodeTo(*c.conn)
}

func (c *Context) HandleRequest(handle func(cc *Context)) {
	handle(c)
}

func (c *Context) ReadResponse() (err error) {
	return c.ResponseConn.DecodeFrom(*c.conn)
}

func (c *Context) SendResponse() (err error) {
	return c.ResponseConn.EncodeTo(*c.conn)
}

func (c *Context) SetRequest(req *protocol.Request) {
	c.RequestConn = req
}

func (c *Context) SetResponse(rsp *protocol.Response) {
	c.ResponseConn = rsp
}
