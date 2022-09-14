package client

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/edte/erpc/log"
	"github.com/edte/erpc/protocol"
	"github.com/edte/erpc/transport"
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
		pooler: transport.NewConnectionPool(),
	}
	return c
}

func (c *Client) Call(ctx context.Context, addr string, req interface{}, rsp interface{}) (err error) {
	// [step 1] 超时处理
	res := &CallRes{
		Done: make(chan struct{}),
		Err:  err,
	}

	go func() {
		c.Do(addr, req, rsp, res)
	}()

	select {
	case <-ctx.Done():
		return errors.New(fmt.Sprintf("call %s timeout", addr))
	case <-res.Done:
		return res.Err
	}
}

func (c *Client) Do(server string, req interface{}, rsp interface{}, res *CallRes) {
	var err error

	defer func() {
		res.Err = err
		res.Done <- struct{}{}
		return
	}()

	log.Debugf("begin call server %s", server)

	// [step 1]  分割 server，格式为 servername.funcname,如果不满足则失败
	s := strings.Split(server, ".")
	if len(s) != 2 {
		err = errors.New("invalid server")
		return
	}

	// alias 下，方便编写
	serverName := s[0]

	log.Debugf("server %s begin discovery", server)

	// [step 2] 先服务发现取目标 ip:port
	addr, err := c.discoveryHttp(serverName)
	if err != nil {
		log.Errorf("server %s discovery http failed, err:%s", serverName, err)
		return
	}

	log.Debugf("server %s discovery succ, res:%s", server, addr)
	log.Debugf("begin get conn from pool, server:%s", server)

	// [step 3] 然后从连接池中取一个连接
	conn, err := c.pooler.GetConn(addr)
	if err != nil {
		log.Errorf("get conn addr %s failed, err:%s", addr, err)
		return
	}

	log.Debugf("get conn from pool succ, server:%s", server)

	// [step 4] 创建连接上下文
	ctx := transport.NewContext(&conn)

	// [step 5] 设置请求协议参数，以及其他上下文参数
	ctx.Request = req
	ctx.Response = rsp
	ctx.RequestConn = protocol.NewRequest(server, req)
	ctx.ResponseConn = &protocol.Response{}
	ctx.ResponseConn.SetEncode(protocol.DefaultBodyCodec)
	ctx.ResponseConn.SetBody(rsp)

	log.Debugf("begin send request, server:%s", server)

	// [setp 6] 发送请求
	if err = ctx.SendRequest(); err != nil {
		return
	}

	log.Debugf("begin read response, server:%s", server)

	// [step 7] 读取响应
	err = ctx.ReadResponse()

	log.Debugf("call server %s succ", server)

	return
}

// 服务发现
// 给定请求服务名，然后负债均衡返回其中一个部署的机器 ip 地址
// TODO: 每次响应一个 ip，那其他集群内怎么同步的？
func (c *Client) discoveryHttp(server string) (addr string, err error) {
	url := "http://127.0.0.1:8080/discovery?server=" + server

	log.Debugf("server %s begin discovery, url:%s", server, url)

	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("server %s get http failed, err:%s", server, err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("server %s read body failed, err:%s", server, err)
		return
	}

	if resp.StatusCode == 500 {
		log.Errorf("server %s response 500, err:%s", server, err)
		return "", errors.New(string(body))
	}

	return string(body), nil
}

func (c *Client) discoveryRpc(server string) (addr string, err error) {
	return
}
