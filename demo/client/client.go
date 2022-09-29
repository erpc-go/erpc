package main

import (
	"context"
	"fmt"
	"time"

	"github.com/edte/testpb2go/demo"
	"github.com/erpc-go/erpc"
)

func main() {
	req := demo.HelloRequest{}
	rsp := demo.HelloResponse{}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)

	if err := erpc.Call(ctx, "demo.hello", &req, &rsp); err != nil {
		panic(fmt.Sprintf("call demo.hello failed, error:%s", err))
	}

	fmt.Println(rsp.Msg)

	req1 := demo.EchoRequest{
		Msg: "ni shi sabe",
	}
	rsp1 := demo.EchoResponse{}

	ctx1, _ := context.WithTimeout(context.Background(), time.Second*5)

	if err := erpc.Call(ctx1, "demo.echo", &req1, &rsp1); err != nil {
		panic(fmt.Sprintf("call demo.echo failed, error:%s", err))
	}

	fmt.Println(rsp1.Msg)
}
