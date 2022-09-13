package server

import "github.com/edte/erpc/transport"

// 异步服务的 handler
type HandlerFunc func(c *transport.Context)

// handler 上下文
type handleItem struct {
	name    string // handle func name
	handler HandlerFunc
	req     interface{}
	rsp     interface{}
}
