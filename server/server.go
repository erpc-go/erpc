package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/edte/erpc/client"
	"github.com/edte/erpc/codec"
	"github.com/edte/erpc/log"
	"github.com/edte/erpc/protocol"
	"github.com/edte/erpc/transport"
	center2 "github.com/edte/testpb2go/center"
)

type RegisterArgs struct {
	Server    string
	Addr      string
	Functions []string
}

// 服务器 server
type Server struct {
	name      string                // server name
	handleMap map[string]handleItem // funcname -> item
	funcList  []string
	localhost string
	args      RegisterArgs
}

func NewServer() *Server {
	s := &Server{
		handleMap: map[string]handleItem{},
		funcList:  []string{},
		args:      RegisterArgs{},
	}

	return s
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

	// [step 4] 心跳
	go s.ping()

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

	s.args = RegisterArgs{
		Server:    s.name,
		Addr:      s.localhost,
		Functions: s.funcList,
	}

	// [setp 1] 先尝试注册一次，成功就直接返回
	err := s.registeHttp()
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
			err := s.registeHttp()
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
func (s *Server) registeHttp() (err error) {
	b, err := json.Marshal(s.args)
	if err != nil {
		return
	}

	resp, err := http.Post("http://127.0.0.1:8080/register", "application/json", strings.NewReader(string(b)))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode == 500 {
		return errors.New(string(body))
	}

	return
}

func (s *Server) registeRpc() (err error) {

	return
}

func (s *Server) handle(conn net.Conn) {
	// [step 1] 循环处理一个连接
	for {
		// log.Debugf("recive a new request")
		// log.Debugf("begin init context")

		// [step 2] 开连接上下文
		c := transport.NewContext(&conn)

		// [step 3] 初始化上下文参数
		c.SetResponseConn(protocol.NewResponse())
		c.RequestConn.SetEncode(codec.CodeTypePb)

		// log.Debugf("begin read request")

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
		log.Debugf("get serve %s handler", host)
		return s.handleMap[host]
	}
	log.Debugf("get handler, name :%s, handleMap raw :%v", st[1], s.handleMap)
	return s.handleMap[st[1]]
}

func (s *Server) ping() {
	for {
		t := time.NewTicker(time.Second)

		select {
		case <-t.C:
			log.Debugf("begin ping center")

			c := client.NewClient()
			req := &center2.HeatRequest{
				SendTime:   time.Now().UnixMilli(),
				ServerName: s.name,
				Addr:       s.localhost,
			}
			err := c.Call(context.Background(), "center.heat", req, &center2.HeatResponse{})
			if err != nil {
				log.Errorf("ping center heat faild,err:%s", err)
				continue
			}

			log.Debugf("ping center succ")
		}
	}
}
