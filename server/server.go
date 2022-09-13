package server

import (
	"io"
	"net"
	"strings"
	"time"

	"github.com/edte/erpc/center"
	"github.com/edte/erpc/codec"
	"github.com/edte/erpc/log"
	"github.com/edte/erpc/protocol"
	"github.com/edte/erpc/transport"
)

// 服务器 server
type Server struct {
	name      string                // server name
	handleMap map[string]handleItem // funcname -> item
	funcList  []string
	localhost string
}

func NewServer() *Server {
	return &Server{
		handleMap: map[string]handleItem{},
		funcList:  []string{},
	}
}

// 本地注册 handler
func (s *Server) Handle(host string, handler HandlerFunc, req interface{}, rsp interface{}) {
	// [step 1] 分割 host
	st := strings.Split(host, ".")
	if len(st) != 2 {
		log.Panicf("invalid name %s", host)
	}

	// [step 2] 设置 servername
	if s.name == "" {
		s.name = st[0]
	} else if s.name != st[0] {
		log.Panicf("multiple service names,%s which should be equal to %s", st[0], s.name)
	}

	// [step 3] 设置 func
	funcName := st[1]
	s.funcList = append(s.funcList, funcName)

	s.handleMap[funcName] = handleItem{
		name:    funcName,
		handler: handler,
		req:     req,
		rsp:     rsp,
	}
}

// 监听服务
func (s *Server) Listen(addr string) {
	log.Debugf("handler:%v", s.handleMap)

	// [step 1] 开 net
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panicf("addr %s listen failed, err:%s", addr, err.Error())
	}

	log.Infof("addr:%s listen succ", addr)

	// [step 2] 设置本地 addr
	s.localhost = addr
	s.localhost = l.Addr().String()

	// [step 3] 服务注册
	go s.register()

	// [step 4] 循环取连接，然后监听
	for {
		c, err := l.Accept()
		if err != nil {
			log.Errorf("server accept faild, addr:%s,err:%s", s.localhost, err)
		}
		go s.handle(c)
	}
}

// 注册服务
// 如果失败则定时 1s 一直注册
func (s *Server) register() {
	done := make(chan struct{}, 1)

	arg := center.ResigerArg{
		Server:    s.name,
		Addr:      s.localhost,
		Functions: s.funcList,
	}

	// [setp 1] 先尝试注册一次，成功就直接返回
	err := center.Register(arg)
	if err == nil {
		log.Infof("serve %s registe succ", arg.Server)
		return
	}

	log.Errorf("server %s registe failed, err:%s", arg.Server, err.Error())

	// [step 2] 失败就定时 1s 重试注册，直到成功
	for {
		t := time.NewTicker(time.Second)

		select {
		case <-done:
			return
		case <-t.C:
			err := center.Register(arg)
			if err == nil {
				log.Infof("serve %s registe succ", arg.Server)
				done <- struct{}{}
				return
			}
			log.Errorf("server %s registe failed, err:%s", arg.Server, err.Error())
		}
	}
}

func (s *Server) handle(conn net.Conn) {
	// [step 1] 循环处理一个连接
	for {
		log.Debugf("recive a new request")
		log.Debugf("begin init context")

		// [step 2] 开连接上下文
		c := transport.NewContext(&conn)

		// [step 3] 初始化上下文参数
		c.SetResponseConn(protocol.NewResponse())
		c.RequestConn.SetEncode(codec.CodeTypePb)

		log.Debugf("begin read request")

		// [step 4] 读请求
		if err := c.ReadRequest(); err != nil {
			if err == io.EOF {
				continue
			}
			log.Errorf("read request failed, err:%s", err)
			c.ResponseConn.SetStatus(protocol.StatusError)
			c.SendResponse()
			continue
		}

		log.Debugf("begin get handler")

		// [step 5] 获取处理 handler item
		h := s.getHandler(c.RequestConn.Host)

		log.Debugf("server %s s'handler :%v", c.RequestConn.Host, h)

		// [step 6] 设置上下文参数 body
		c.RequestConn.SetBody(h.req)
		c.ResponseConn.SetBody(h.rsp)
		c.Request = h.req
		c.Response = h.rsp

		log.Debugf("begin read request body")

		// [step 7] 解析请求体
		if err := c.ReadRequestBody(); err != nil {
			log.Errorf("read request body failed")
			c.ResponseConn.SetStatus(protocol.StatusError)
			c.SendResponse()
			continue
		}

		log.Debugf("begin handle request")

		// [step 8] 处理请求
		c.HandleRequest(h.handler)

		log.Debugf("begin send response")

		// [step 9] 发送响应
		c.SendResponse()

		log.Debugf("send response succ")
	}
}

func (s *Server) getServerNameByHost(host string) string {
	st := strings.Split(host, ".")
	if len(st) != 2 {
		log.Errorf("invalid host %s", host)
	}
	return st[0]
}

func (s *Server) setServerName(host string) {

}

func (s *Server) getHandler(host string) handleItem {
	st := strings.Split(host, ".")
	if len(st) != 2 {
		return s.handleMap[host]
	}
	return s.handleMap[st[1]]
}
