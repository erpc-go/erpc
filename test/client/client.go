package main

import (
	"fmt"

	"github.com/edte/erpc"
	"github.com/edte/testpb2go/ping"
)

func main() {
	req := ping.PingRequest{
		Msg: "ping",
	}
	rsp := ping.PingResponse{}

	if err := erpc.Call("demo.ping", &req, &rsp); err != nil {
		panic(fmt.Sprintf("call demo.ping failed, error:%s", err))
	}

	fmt.Println(rsp.Msg)
}
