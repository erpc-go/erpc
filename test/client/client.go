package main

import (
	"fmt"

	"github.com/edte/erpc"
	"github.com/edte/testpb2go/echo"
	"github.com/edte/testpb2go/hello"
)

func main() {
	req := hello.HelloRequest{}
	rsp := hello.HelloResponse{}

	if err := erpc.Call("demo.hello", &req, &rsp); err != nil {
		panic(fmt.Sprintf("call demo.hello failed, error:%s", err))
	}

	fmt.Println(rsp.Msg)

	req1 := echo.EchoRequest{
		Msg: "ni shi sabe",
	}
	rsp1 := echo.EchoResponse{}

	if err := erpc.Call("demo.echo", &req1, &rsp1); err != nil {
		panic(fmt.Sprintf("call demo.echo failed, error:%s", err))
	}

	fmt.Println(rsp1.Msg)
}
