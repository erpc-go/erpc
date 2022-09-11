package main

import (
	"fmt"

	"github.com/edte/erpc"
	"github.com/edte/erpc/transport"
	"github.com/edte/testpb2go/echo"
	"github.com/edte/testpb2go/hello"
)

func hanleHello(c *transport.Context) {
	req := c.Request.(*hello.HelloRequest)
	rsp := c.Response.(*hello.HelloResponse)

	req.Msg = "hello"
	fmt.Println(rsp.Msg)
}

func hanleEcho(c *transport.Context) {
	req := c.Request.(*echo.EchoRequest)
	rsp := c.Response.(*echo.EchoResponse)

	req.Msg = "hello"
	fmt.Println(rsp.Msg)
}

func main() {
	erpc.Handle("hello", hanleHello, &hello.HelloRequest{}, &hello.HelloResponse{})
	erpc.Handle("echo", hanleEcho, &echo.EchoRequest{}, &echo.EchoResponse{})
	erpc.Listen(":8080")
}
