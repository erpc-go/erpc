package client

import (
	"errors"
	"strings"
	"time"

	"github.com/edte/erpc/center"
	"github.com/edte/erpc/log"
	"github.com/edte/erpc/protocol"
	"github.com/edte/erpc/transport"
)

type Client struct {
	pooler         transport.Pooler // 底层连接池
	ConnectTimeout time.Duration    // 连接超时设置
}

func NewClient() *Client {
	c := &Client{
		pooler: &transport.ConnectionPool{},
	}
	return c
}

// TODO: 请求超时设置
func (c *Client) Call(server string, req interface{}, rsp interface{}) (err error) {
	log.Debugf("begin call server %s", server)

	// [step 1]  分割 server，格式为 servername.funcname,如果不满足则失败
	s := strings.Split(server, ".")
	if len(s) != 2 {
		return errors.New("invalid server")
	}

	// alias 下，方便编写
	serverName := s[0]

	log.Debugf("server %s begin discovery", server)

	// [step 2] 先服务发现取目标 ip:port
	addr, err := center.Discovery(serverName)
	if err != nil {
		return
	}

	log.Debugf("server %s discovery succ, res:%s", server, addr)
	log.Debugf("begin get conn from pool, server:%s", server)

	// [step 3] 然后从连接池中取一个连接
	conn, err := c.pooler.GetConn(addr)
	if err != nil {
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
	return ctx.ReadResponse()
}
