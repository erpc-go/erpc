package server

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"time"

	"github.com/edte/erpc/center/contant"
	"github.com/edte/erpc/client"
	"github.com/edte/erpc/codec"
	"github.com/edte/erpc/log"
	"github.com/edte/erpc/protocol"
	"github.com/edte/erpc/transport"
	"github.com/edte/testpb2go/center"
	center2 "github.com/edte/testpb2go/center"
)

// 服务器 server
type Server struct {
	name      string  // server name
	handles   handles // funcname -> item
	funcs     []string
	localhost string

	iscenter bool
}

func NewServer() *Server {
	s := &Server{
		handles:  map[string]handleItem{},
		funcs:    []string{},
		iscenter: false,
	}

	return s
}

func (s *Server) SetCenter() {
	s.iscenter = true
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
	s.funcs = append(s.funcs, funcName)

	s.handles.add(funcName, handler, req, rsp)
}

// 监听服务
func (s *Server) Listen(add string) {
	addr := add[5:]

	log.Debugf("handles:%v", s.handles)

	// [step 1] 开 net
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panicf("addr %s listen failed, err:%s", addr, err.Error())
	}

	log.Infof("addr:%s listen succ", addr)

	// [step 2] 设置本地 addr
	s.localhost = addr
	s.localhost = l.Addr().String()

	// [step 3] 如果不是 center server，则开始注册服务和心跳
	if !s.iscenter {
		go s.register()
		go s.heatbeat()
	}

	// [step 5] 循环取连接，然后监听
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

	// [setp 1] 先尝试注册一次，成功就直接返回
	err := s.registe()
	if err == nil {
		log.Infof("serve %s registe succ", s.name)
		return
	}

	log.Errorf("server %s registe failed, err:%s", s.name, err.Error())

	// [step 2] 失败就定时 1s 重试注册，直到成功
	for {
		t := time.NewTicker(time.Second)

		select {
		case <-done:
			return
		case <-t.C:
			err := s.registe()
			if err == nil {
				log.Infof("serve %s registe succ", s.name)
				done <- struct{}{}
				return
			}
			log.Errorf("server %s registe failed, err:%s", s.name, err.Error())
		}
	}
}

// TODO: 考虑本地服务缓存提供服务发现、服务注册功能？（去注册中心化？考虑优化）
func (s *Server) registe() (err error) {
	req := &center.RegisterRequest{
		ServerName: s.name,
		Addr:       s.localhost,
		Functions:  s.funcs,
	}
	rsp := &center.RegisterResponse{}

	c := client.NewClient()

	if err = c.Call(context.Background(), contant.RouteRegister, req, rsp); err != nil {
		log.Errorf("server %s register failed,err:%s", s.name, err)
		return
	}

	return errors.New(rsp.Err)
}

func (s *Server) handle(conn net.Conn) {
	// [step 1] 循环处理一个连接
	for {
		log.Debugf("recive a new request")
		log.Debugf("begin init context")

		// [step 2] 开连接上下文
		c := transport.NewContext()
		c.SetConn(conn)

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

		log.Debugf("begin get handler, request:%v", c.RequestConn)

		// [step 5] 获取处理 handler item
		h := s.handles.get(c.RequestConn.Route)

		log.Debugf("server %s s'handler :%v", c.RequestConn.Route, h)

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

// FIX: center 也是一个 server，那么 center server 也需要向 center 发心跳？（原地原地是吧)
func (s *Server) heatbeat() {
	req := &center2.HeatRequest{
		SendTime:   time.Now().UnixMilli(),
		ServerName: s.name,
		Addr:       s.localhost,
	}
	rsp := &center2.HeatResponse{}
	c := client.NewClient()

	for {
		t := time.NewTicker(time.Second)

		select {
		case <-t.C:
			log.Debugf("server %s begin heatbeat center", s.name)

			req.SendTime = time.Now().UnixMilli()

			err := c.Call(context.Background(), "center.heat", req, rsp)
			if err != nil {
				log.Errorf("heatbeat center heat faild,err:%s", err)
				continue
			}

			log.Debugf("heatbeat center succ, recive time:%s", time.UnixMilli(rsp.ResponseTime))
		}
	}
}
