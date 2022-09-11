package server

import (
	"fmt"
	"net"

	"github.com/edte/erpc/center"
	"github.com/edte/erpc/protocol"
	"github.com/edte/erpc/transport"
)

// 异步服务的 handler
type HandlerFunc func(c *transport.Context)

// handler 上下文
type handleItem struct {
	partten string
	handler HandlerFunc
	req     interface{}
	rsp     interface{}
}

// 服务器 server
type Server struct {
	handleMap map[string]handleItem
	localhost string
}

func NewServer() *Server {
	return &Server{
		handleMap: map[string]handleItem{},
	}
}

// 本地注册 handler
func (s *Server) Handle(partten string, handler HandlerFunc, req interface{}, rsp interface{}) {
	s.handleMap[partten] = handleItem{
		partten: partten,
		handler: handler,
		req:     req,
		rsp:     rsp,
	}
}

// 监听服务
func (s *Server) Listen(addr string) {
	// [step 1] 开 net
	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	// [step 2] 设置本地 addr
	s.localhost = l.Addr().String()

	// [step 3] 服务注册
	for _, h := range s.handleMap {
		if err := center.Register(h.partten, s.localhost, h.req, h.rsp); err != nil {
			panic(err)
		}
	}

	// [step 4] 循环取连接，然后监听
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Printf("accept faild, addr:%s,err:%s", s.localhost, err)
		}
		go s.handle(c)
	}
}

func (s *Server) handle(conn net.Conn) {
	// [step 1] 循环处理一个连接
	for {
		// [step 2] 开连接上下文
		c := transport.NewContext(&conn)

		// [step 3] 读请求
		if err := c.ReadRequest(); err != nil {
			c.ResponseConn.SetStatus(protocol.StatusError)
			c.SendResponse()
			continue
		}

		// [step 4] 处理请求
		c.HandleRequest(s.handleMap[c.RequestConn.Host].handler)

		// [step 5] 发送响应
		c.SendResponse()
	}
}
