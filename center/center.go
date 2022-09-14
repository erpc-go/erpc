package center

import (
	"time"

	"github.com/edte/erpc/log"
	"github.com/edte/erpc/server"
	"github.com/edte/erpc/transport"
	"github.com/edte/testpb2go/center"
	"github.com/gin-gonic/gin"
)

type RegisterArgs struct {
	Server    string
	Addr      string
	Functions []string
}

var (
	DefaultCenterHttpAddress = "127.0.0.1:8080"
	DefaultCenterRpcAddress  = "127.0.0.1:8081"
)

// TODO: 把负载均衡功能拆分出来，然后服务发现的时候只返回 l5 地址，而不是直接是 ip，把选的过程拆分到 l5，让 client 去调用

// 注册中心
// 提供服务注册、服务发现功能
type Center struct {
	listenAddr string
	servers    servers
}

func NewCenter() *Center {
	c := &Center{
		servers: map[string]*serverItem{},
	}
	return c
}

func (c *Center) Listen() {
	go c.listenHttp()
	go c.updateServer()
	c.listenRpc()
}

func (c *Center) listenRpc() {
	s := server.NewServer()
	s.Handle("center.discovery", c.rpcDiscoverHandler(), &center.DiscoveryRequest{}, &center.DiscoveryResponse{})
	s.Handle("center.register", c.rpcRegisterHandler(), &center.RegisterRequest{}, &center.RegisterResponse{})
	s.Handle("center.heat", c.rpcPingHandler(), &center.HeatRequest{}, &center.HeatResponse{})
	s.Listen(DefaultCenterRpcAddress)
}

func (c *Center) listenHttp() {
	r := gin.Default()
	r.GET("/discovery", c.httpDiscoveryHandler())
	r.POST("/register", c.httpRegisterHandler())
	r.Run(DefaultCenterHttpAddress)
}

func (c *Center) rpcDiscoverHandler() server.HandlerFunc {
	return func(ctx *transport.Context) {
		req := ctx.Request.(*center.DiscoveryRequest)
		rsp := ctx.Response.(*center.DiscoveryResponse)

		log.Debugf("begin recevie rpc discover")

		res, err := c.Discovery(req.Server)
		if err != nil {
			log.Errorf("server %s discovery failed, err:%s", req.Server, err)
			rsp.Err = err.Error()
			return
		}

		log.Debugf("serve %s discover succ, res:%s", req.Server, res)
		rsp.Err = ""
	}
}

func (c *Center) rpcRegisterHandler() server.HandlerFunc {
	return func(ctx *transport.Context) {
		req := ctx.Request.(*center.RegisterRequest)
		rsp := ctx.Response.(*center.RegisterResponse)

		log.Debugf("begin deal rpc register")

		r := &RegisterArgs{
			Server:    req.ServerName,
			Addr:      req.Addr,
			Functions: req.Functions,
		}

		if err := c.Register(r.Server, r.Addr, r.Functions); err != nil {
			log.Errorf("registe failed, err:%s", err)
			rsp.Err = err.Error()
			return
		}

		rsp.Err = ""
		log.Debugf("register %s succ", r.Server)
	}
}

func (c *Center) rpcPingHandler() server.HandlerFunc {
	return func(ctx *transport.Context) {
		req := ctx.Request.(*center.HeatRequest)
		rsp := ctx.Response.(*center.HeatResponse)

		s := c.servers[req.ServerName]
		s.heatbeat(req.Addr, req.SendTime)

		rsp.ResponseTime = time.Now().UnixNano()
	}
}

func (c *Center) httpDiscoveryHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		s := ctx.Query("server")
		res, err := c.Discovery(s)
		if err != nil {
			log.Errorf("server %s discovery failed, err:%s", s, err)
			ctx.String(500, err.Error())
			return
		}
		log.Debugf("serve %s discover succ, res:%s", s, res)
		ctx.String(200, res)
	}
}

func (c *Center) httpRegisterHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		r := &RegisterArgs{}

		log.Debugf("recevie register http")

		if err := ctx.BindJSON(&r); err != nil {
			log.Errorf("registe failed, err:%s", err)
			ctx.String(500, err.Error())
			return
		}

		log.Debugf("args:%v", r)

		if err := c.Register(r.Server, r.Addr, r.Functions); err != nil {
			log.Errorf("registe failed, err:%s", err)
			ctx.String(500, err.Error())
			return
		}

		log.Debugf("register %s succ", r.Server)
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
