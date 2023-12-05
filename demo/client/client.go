package main

import (
	"context"
	"fmt"

	"github.com/erpc-go/erpc"
	"github.com/erpc-go/testjce2go/base"
)

func main() {
	req := base.AndroidReq{
		StrMessageId: "test",
	}
	rsp := base.AndroidRsp{}

	if err := erpc.Do(context.Background(), "test", &req, &rsp); err != nil {
		panic(fmt.Sprintf("call failed, error:%s", err))
	}

	fmt.Println(rsp)
}
