package center

import (
	"strings"
	"time"

	"github.com/edte/testpb2go/center"
	"github.com/erpc-go/erpc/center/contant"
	"github.com/erpc-go/erpc/server"
	"github.com/erpc-go/erpc/transport"
	"github.com/erpc-go/log"
)

type RegisterArgs struct {
	Server    string
	Addr      string
	Functions []string
}

// TODO: 增加多机房，路由功能

// TODO: 把负载均衡功能拆分出来，然后服务发现的时候只返回 l5 地址，而不是直接是 ip，把选的过程拆分到 l5，让 client 去调用

// 注册中心
// 提供服务注册、服务发现功能
type Center struct {
	servers servers
}

func NewCenter() *Center {
	c := &Center{
		servers: map[string]*serverItem{},
	}
	return c
}

func (c *Center) Listen() {
	go c.updateServer()

	s := server.NewServer()
	s.SetCenter()
	s.Handle(contant.RouteDiscovery, c.HandlerDiscover(), &center.DiscoveryRequest{}, &center.DiscoveryResponse{})
	s.Handle(contant.RouteRegister, c.HandlerRegister(), &center.RegisterRequest{}, &center.RegisterResponse{})
	s.Handle(contant.RouteHeatbeat, c.HandlerHeat(), &center.HeatRequest{}, &center.HeatResponse{})
	s.Listen(contant.DefaultCenterAddress)
}

func (c *Center) HandlerDiscover() server.HandlerFunc {
	return func(ctx *transport.Context) {
		req := ctx.Request.(*center.DiscoveryRequest)
		rsp := ctx.Response.(*center.DiscoveryResponse)

		log.Debug("center begin recevie discover %s", req.Server)

		server := ""

		s := strings.Split(req.Server, ".")
		if len(s) == 1 {
			server = req.Server
		} else if len(s) == 2 {
			server = s[0]
		} else {
			rsp.Err = "server format invalid"
			return
		}

		log.Debug("center ,server name is %s", server)

		addr, err := c.Discovery(server)
		if err != nil {
			log.Error("center, server %s discovery failed, err:%s", req.Server, err)
			rsp.Err = err.Error()
			return
		}

		log.Debug("center, serve %s discover succ, res:%s", req.Server, addr)

		rsp.Err = ""
		rsp.Addr = addr
	}
}

func (c *Center) HandlerRegister() server.HandlerFunc {
	return func(ctx *transport.Context) {
		req := ctx.Request.(*center.RegisterRequest)
		rsp := ctx.Response.(*center.RegisterResponse)

		log.Debug("center begin deal rpc register")

		r := &RegisterArgs{
			Server:    req.ServerName,
			Addr:      req.Addr,
			Functions: req.Functions,
		}

		if err := c.Register(r.Server, r.Addr, r.Functions); err != nil {
			log.Error("center registe failed, err:%s", err)
			rsp.Err = err.Error()
			return
		}

		log.Debug("center register %s succ", r.Server)
	}
}

// asdfasdfasdf

func (c *Center) HandlerHeat() server.HandlerFunc {
	return func(ctx *transport.Context) {
		req := ctx.Request.(*center.HeatRequest)
		rsp := ctx.Response.(*center.HeatResponse)

		s := c.servers[req.ServerName]
		s.heatbeat(req.Addr, req.SendTime)

		rsp.ResponseTime = time.Now().UnixNano()
	}
}

func (c *Center) Register(server string, addr string, funcs []string) (err error) {
	return c.servers.registe(server, addr, funcs)
}

func (c *Center) Discovery(server string) (addr string, err error) {
	return c.servers.discovery(server)
}

// 周期 update 失效的 server
func (c *Center) updateServer() {
	c.servers.update()
}
