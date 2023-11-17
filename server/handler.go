package server

import (
	"strings"

	"github.com/erpc-go/erpc/log"
	"github.com/erpc-go/erpc/transport"
)

// 异步服务的 handler
type HandlerFunc func(c *transport.Context)

// handler 上下文
type handleItem struct {
	name    string // handle func name
	handler HandlerFunc
	req     interface{}
	rsp     interface{}
}

// handle list
type handles map[string]handleItem // funcname -> item

func (h handles) add(name string, handler HandlerFunc, req interface{}, rsp interface{}) {
	h[name] = handleItem{
		name:    name,
		handler: handler,
		req:     req,
		rsp:     rsp,
	}
}

func (h handles) delete(name string) {
	delete(h, name)
}

func (h handles) update(name string, handler HandlerFunc, req interface{}, rsp interface{}) {
	h.add(name, handler, req, rsp)
}

// 支持 servername.funcname 或者 funcname 两种访问方式
func (h handles) get(name string) (res handleItem, has bool) {
	server := name

	st := strings.Split(name, ".")
	if len(st) == 2 {
		server = st[1]
	}

	log.Debug("get serve %s handler", name)
	log.Debug("get handler,name :%s,handles raw :%v", st[1], h)

	res, has = h[server]
	return
}
