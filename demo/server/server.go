package main

import (
	"fmt"

	"github.com/edte/testpb2go/demo"
	"github.com/erpc-go/erpc"
	"github.com/erpc-go/erpc/transport"
	"github.com/erpc-go/log"
)

func handleHello(c *transport.Context) {
	rsp := c.Response.(*demo.HelloResponse)

	rsp.Msg = "hello world"
	fmt.Println(rsp.Msg)
}

func handleEcho(c *transport.Context) {
	req := c.Request.(*demo.EchoRequest)
	rsp := c.Response.(*demo.EchoResponse)

	rsp.Msg = req.Msg

	fmt.Println(rsp.Msg)
}

func main() {
	log.DefaultLogger.SetLevel(log.DebugLevel)
	erpc.Handle("demo.hello", handleHello, &demo.HelloRequest{}, &demo.HelloResponse{})
	erpc.Handle("demo.echo", handleEcho, &demo.EchoRequest{}, &demo.EchoResponse{})
	erpc.Listen(":8877")
}
