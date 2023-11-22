package main

import (
	"fmt"

	"github.com/erpc-go/erpc"
	"github.com/erpc-go/erpc/server"
	"github.com/erpc-go/log"
	"github.com/erpc-go/testjce2go/base"
)

func handleHello(c *server.Context) {
	rsp := c.Rsp.(*base.AndroidRsp)
	fmt.Println(rsp)
}

func main() {
	log.DefaultLogger.SetLevel(log.DebugLevel)
	erpc.GlobalServeMutex.ListenNet = "tcp"
	erpc.GlobalServeMutex.Address = ":888"
	erpc.HandleFunc("demo.test.hh.send", "", handleHello, (*base.AndroidReq)(nil), (*base.AndroidRsp)(nil))
	erpc.ListenAndServe()
}
