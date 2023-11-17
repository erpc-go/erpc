package server

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"time"

	"github.com/edte/testpb2go/center"
	center2 "github.com/edte/testpb2go/center"
	"github.com/erpc-go/erpc/center/contant"
	"github.com/erpc-go/erpc/client"
	"github.com/erpc-go/erpc/codec"
	"github.com/erpc-go/erpc/log"
	"github.com/erpc-go/erpc/protocol"
	"github.com/erpc-go/erpc/transport"
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
		log.Panic("invalid name %s", host)
	}

	// [step 2] 设置 servername
	if s.name == "" {
		s.name = st[0]
	} else if s.name != st[0] {
		log.Panic("multiple service names,%s which should be equal to %s", st[0], s.name)
	}

	// [step 3] 设置 func
	funcName := st[1]
	s.funcs = append(s.funcs, funcName)

	s.handles.add(funcName, handler, req, rsp)
}

// 监听服务
func (s *Server) Listen(addr string) {
	log.Debug("handles:%v", s.handles)

	// [step 1] 开 net
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panic("addr %s listen failed, err:%s", addr, err.Error())
	}

	log.Info("addr:%s listen succ", addr)

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
			log.Error("server accept faild, addr:%s,err:%s", s.localhost, err)
		}

		log.Debug("server %s accept a new client:%s", addr, c.RemoteAddr())

		// TODO: 这里考虑优化掉，当连接断开后，协程也关闭，而不是继续循环
		go s.handle(c)
	}
}

// 注册服务
// 如果失败则定时 1s 一直注册
func (s *Server) register() {
	done := make(chan struct{}, 1)

	log.Debug("server %s begin first registe", s.name)

	// [setp 1] 先尝试注册一次，成功就直接返回
	err := s.registe()
	if err == nil {
		log.Info("serve %s first registe succ", s.name)
		return
	}

	log.Error("server %s first registe failed, err:%s", s.name, err.Error())

	// [step 2] 失败就定时 1s 重试注册，直到成功
	for {
		t := time.NewTicker(time.Second)

		select {
		case <-done:
			return
		case <-t.C:
			log.Debug("server %s begin registe", s.name)

			err := s.registe()

			log.Debug("raw xxxhh:%v, %v", err, err == nil)

			if err == nil {
				log.Info("serve %s registe succ", s.name)
				done <- struct{}{}
				return
			}

			log.Error("server %s registe failed, err:%s", s.name, err.Error())
		}
	}
}

// TODO: 考虑本地服务缓存提供服务发现、服务注册功能？（去注册中心化？考虑优化）
func (s *Server) registe() error {
	req := &center.RegisterRequest{
		ServerName: s.name,
		Addr:       s.localhost,
		Functions:  s.funcs,
	}
	rsp := &center.RegisterResponse{}

	c := client.NewClient()

	if err := c.Call(context.Background(), contant.RouteRegister, req, rsp); err != nil {
		log.Error("server %s register failed,err:%s", s.name, err)
		return err
	}

	// log.Debugf("registe response:%v", rsp.Err)
	if rsp.Err != "" {
		return errors.New(rsp.Err)
	}

	return nil
}

func (s *Server) handle(conn net.Conn) {
	// [step 1] 循环处理一个连接
	for {
		log.Debug("%s recive a new request", s.name)
		log.Debug("%s begin init context", s.name)

		// [step 2] 开连接上下文
		c := transport.NewContext()
		c.SetConn(conn)

		log.Debug("first %s context raw:%v", s.name, c)

		// [step 3] 初始化上下文参数
		c.SetResponseConn(protocol.NewResponse())
		c.RequestConn.SetEncode(codec.CodeTypePb)

		log.Debug("%s begin read request", s.name)

		// [step 4] 读请求
		// FIX: 这里注意，现有逻辑是 read 失败会直接返回，理论上应该是阻塞而不是返回
		// net 的 read 操作是阻塞的，但是不知道为啥变非阻塞了
		if err := c.ReadRequest(); err != nil {
			if err == io.EOF {
				continue
			}
			log.Error("%s read request failed, err:%s", s.name, err)
			c.ResponseConn.SetStatus(protocol.StatusError)
			c.SendResponse()
			continue
		}

		log.Debug("%s begin get handler, request:%v", s.name, c.RequestConn)

		// [step 5] 获取处理 handler item
		h, has := s.handles.get(c.RequestConn.Route)
		if !has {
			log.Debug("server %s has not registe %s", s.name, c.RequestConn.Route)
			continue
		}

		log.Debug("%s s'handler :%v", c.RequestConn.Route, h)

		// [step 6] 设置上下文参数 body
		c.RequestConn.SetBody(h.req)
		c.ResponseConn.SetBody(h.rsp)
		c.Request = h.req
		c.Response = h.rsp

		log.Debug("second %s context raw:%v", s.name, c)

		log.Debug("%s begin read request body", s.name)

		// [step 7] 解析请求体
		if err := c.ReadRequestBody(); err != nil {
			log.Error("%s read request body failed", s.name)
			c.ResponseConn.SetStatus(protocol.StatusError)
			c.SendResponse()
			continue
		}

		log.Debug("%s begin handle request", s.name)

		// [step 8] 处理请求
		c.HandleRequest(h.handler)

		log.Debug("%s begin send response", s.name)

		// [step 9] 发送响应
		c.SendResponse()

		log.Debug("%s send response succ", s.name)
	}
}

// FIX: center 也是一个 server，那么 center server 也需要向 center 发心跳？（原地原地是吧)
// FIX: 这里心跳的时候，每次发都是一个新的 socket 连接，开销大（考虑连接池优化）
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
			log.Debug("server %s begin heatbeat center", s.name)

			req.SendTime = time.Now().UnixMilli()

			// TODO: 心跳包考虑用 UDP？

			err := c.Call(context.Background(), contant.RouteHeatbeat, req, rsp)
			if err != nil {
				log.Error("heatbeat center heat faild,err:%s", err)
				continue
			}

			log.Debug("heatbeat center succ, recive time:%s", time.UnixMilli(rsp.ResponseTime))
		}
	}
}
