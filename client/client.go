package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/edte/erpc/center/contant"
	"github.com/edte/erpc/log"
	"github.com/edte/erpc/protocol"
	"github.com/edte/erpc/transport"
	"github.com/edte/testpb2go/center"
)

var (
	DefaultConnectionPool transport.Pooler = transport.NewConnectionPool()
)

type CallRes struct {
	Done chan struct{}
	Err  error
}

type Client struct {
	pooler         transport.Pooler // 底层连接池
	ConnectTimeout time.Duration    // 连接超时设置
}

func NewClient() *Client {
	c := &Client{
		pooler: DefaultConnectionPool,
	}
	return c
}

// 对外的接口 1： 参数自动打包，直接传特定参数
func (c *Client) Call(ctx context.Context, route string, req interface{}, rsp interface{}) (err error) {
	log.Debugf("client begin init context when request addr:%s", route)

	// [step 1] 创建连接上下文
	ct := transport.NewContext()

	// [step 2] 设置请求协议参数，以及其他上下文参数
	ct.Request = req
	ct.Response = rsp

	ct.RequestConn = protocol.NewRequest(route, req)
	ct.ResponseConn = &protocol.Response{}
	ct.ResponseConn.SetEncode(protocol.DefaultBodyCodec)
	ct.ResponseConn.SetBody(rsp)

	// [step 3] 开始
	return c.listen(ctx, ct)
}

// 对外的接口 2：上下文手动写，所有东西都自己填
func (c *Client) Do(ctx context.Context, ct *transport.Context) (err error) {
	return c.listen(ctx, ct)
}

func (c *Client) listen(ctx context.Context, ct *transport.Context) (err error) {
	res := &CallRes{
		Done: make(chan struct{}),
	}

	// [step 1] 开始发送
	go func() {
		log.Debugf("client begin send,raw:%v", ct)
		if err := c.send(ct, res); err != nil {
			log.Errorf("client send failed, raw ctx:%v", ct)
			return
		}
	}()

	// [step 2] 设置超时
	select {
	case <-ctx.Done():
		return errors.New(fmt.Sprintf("send %s timeout", ct.RequestConn.Route))
	case <-res.Done:
		return res.Err
	}
}

func (c *Client) send(ctx *transport.Context, res *CallRes) (err error) {
	defer func() {
		res.Err = err
		res.Done <- struct{}{}
		return
	}()

	route := ctx.RequestConn.Route

	log.Debugf("begin get addr %s when send", route)

	// [step 1] 获取 ip:port
	addr, err := c.getAddr(route)
	if err != nil {
		log.Errorf("server %s get addr failed, err:%s", route, err)
		return
	}

	log.Debugf("begin get conn from pool, server:%s, addr:%s", route, addr)

	// [step 2] 然后从连接池中取一个连接
	conn, err := c.pooler.GetConn(addr)
	if err != nil {
		log.Errorf("get conn addr %s failed, err:%s", addr, err)
		return
	}

	// [setp 3] 设置 context 连接
	ctx.SetConn(conn)

	log.Debugf("begin init connection context when request server: %s", route)

	// [setp 4] 发送请求
	if err = ctx.SendRequest(); err != nil {
		return
	}

	log.Debugf("begin read response, server:%s", route)

	// [step 5] 读取响应
	err = ctx.ReadResponse()

	log.Debugf("call server %s succ", route)

	return
}

// 支持两种格式
// 1. ip://
// 2. servername.funcname
func (c *Client) getAddr(server string) (addr string, err error) {
	log.Debugf("client begin get server %s 's addr", server)

	// [step 1] 如果是 ip，就直接走
	ip, isIp := c.getIp(server)
	if isIp {
		log.Debugf("%s is ip", server)
		return ip, nil
	}

	log.Debugf("%s is servername.funcname type", server)

	t := server

	// [step 2]  分割 server，格式为 servername.funcname, 如果不满足则失败
	s := strings.Split(server, ".")
	if len(s) == 1 {
		t = server
	} else if len(s) == 2 {
		t = s[0]
	} else {
		log.Errorf("invalid server format, neet [servername.funcname] type, route:%s", server)
		err = errors.New("invalid server format, neet [servername.funcname] type")
		return
	}

	// [step 3] 如果是 servername 格式，就去中心服务发现
	return c.discovery(t)
}

func (c *Client) isIp(server string) bool {
	_, isIp := c.getIp(server)
	return isIp
}

// 格式校验例子如下：
// ip://127.0.0.1:8080
func (c *Client) getIp(server string) (ip string, isIp bool) {
	// [step 1] 先校验长度
	if len(server) <= 5 {
		log.Debugf("server %s 'length <= 5", server)
		return "", false
	}

	// [step 2] 然后取 schema 和后面的
	schema := server[:5]
	last := server[5:]

	log.Debugf("schema: %s, last: %s", schema, last)

	// [step 3] 校验协议
	if schema != "ip://" {
		log.Debugf("shema %s not equal ip://", schema)
		return "", false
	}

	// [step 3] 再分离 ip 和 port
	s := strings.Split(last, ":")
	if len(s) != 2 {
		log.Debugf("split %s error by :", last)
		return "", false
	}

	i, err := strconv.Atoi(s[1])
	if err != nil {
		log.Debugf("port %s not number", s[1])
		return "", false
	}

	// [step 4] 最后总的校验 ip 和端口范围
	isIp = net.ParseIP(s[0]) != nil && i >= 0 && i <= 65535
	if isIp {
		return last, true
	}
	return "", false
}

// 服务发现
// 给定请求服务名，然后负债均衡返回其中一个部署的机器 ip 地址
// TODO: 每次响应一个 ip，那其他集群内怎么同步的？
func (c *Client) discovery(route string) (addr string, err error) {
	log.Debugf("begin discovery %s", route)

	if route == "center" {
		return contant.DefaultCenterAddress, nil
	}

	req := &center.DiscoveryRequest{
		Server: route,
	}
	rsp := &center.DiscoveryResponse{}

	if err = c.Call(context.Background(), contant.RouteDiscovery, req, rsp); err != nil {
		log.Errorf("client discovery %s call failed, err:%s", route, err)
		return
	}

	if rsp.Err != "" {
		log.Errorf("client discovery %s server failed, err:%s", route, rsp.Err)
		return "", errors.New(rsp.Err)
	}

	log.Debugf("discovery route %s succ,res:%s", route, rsp.Addr)

	addr = rsp.Addr

	return
}
