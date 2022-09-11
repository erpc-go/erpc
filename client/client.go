package client

import (
	"time"

	"github.com/edte/erpc/center"
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
	// [step 1] 先服务发现取目标 ip:port
	addr, err := center.Discovery(server)
	if err != nil {
		return
	}

	// [step 2] 然后从连接池中取一个连接
	conn, err := c.pooler.GetConn(addr)
	if err != nil {
		return
	}

	// [step 3] 创建连接上下文
	ctx := transport.NewContext(&conn)
	ctx.Request = req
	ctx.Response = rsp

	// [setp 4] 发送请求
	if err = ctx.SendRequest(); err != nil {
		return
	}

	// [step 5] 读取响应
	if err = ctx.ReadResponse(); err != nil {
		return
	}

	return
}
