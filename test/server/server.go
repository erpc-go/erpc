package main

import (
	"fmt"

	"github.com/edte/erpc"
	"github.com/edte/erpc/transport"
	"github.com/edte/testpb2go/echo"
	"github.com/edte/testpb2go/hello"
)

func handleHello(c *transport.Context) {
	rsp := c.Response.(*hello.HelloResponse)

	rsp.Msg = "hello world"
	fmt.Println(rsp.Msg)
}

func handleEcho(c *transport.Context) {
	req := c.Request.(*echo.EchoRequest)
	rsp := c.Response.(*echo.EchoResponse)

	rsp.Msg = req.Msg

	fmt.Println(rsp.Msg)
}

func main() {
	erpc.Handle("demo.hello", handleHello, &hello.HelloRequest{}, &hello.HelloResponse{})
	erpc.Handle("demo.echo", handleEcho, &echo.EchoRequest{}, &echo.EchoResponse{})
	erpc.Listen(":8877")
}
