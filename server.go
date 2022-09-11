package erpc

import "github.com/edte/erpc/server"

var (
	defaultServer = server.NewServer()
)

func Handle(partten string, handler server.HandlerFunc, req interface{}, rsp interface{}) {
	defaultServer.Handle(partten, handler, req, rsp)
}

func Listen(addr string) {
	defaultServer.Listen(addr)
}
