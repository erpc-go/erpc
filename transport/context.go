package transport

import (
	"github.com/edte/erpc/protocol"
)

// 连接上下文
type Context struct {
	RequestConn  *protocol.Request
	ResponseConn *protocol.Response

	Request  interface{}
	Response interface{}

	conn Conn
}

func NewContext() *Context {
	return &Context{
		RequestConn:  &protocol.Request{},
		ResponseConn: &protocol.Response{},
	}
}

func (c *Context) SetConn(conn Conn) {
	c.conn = conn
}

// set
func (c *Context) SetRequestConn(req *protocol.Request) {
	c.RequestConn = req
}

func (c *Context) SetResponseConn(rsp *protocol.Response) {
	c.ResponseConn = rsp
}

// 解码整个请求
func (c *Context) ReadRequest() (err error) {
	return c.RequestConn.DecodeFrom(c.conn)
}

// 解码请求 body
func (c *Context) ReadRequestBody() (err error) {
	return c.RequestConn.DecodeBody()
}

// 发送请求
func (c *Context) SendRequest() (err error) {
	return c.RequestConn.EncodeTo(c.conn)
}

// 处理请求
func (c *Context) HandleRequest(handle func(cc *Context)) {
	handle(c)
}

// 读取响应
func (c *Context) ReadResponse() (err error) {
	return c.ResponseConn.DecodeFrom(c.conn)
}

// 发送响应
func (c *Context) SendResponse() (err error) {
	return c.ResponseConn.EncodeTo(c.conn)
}
