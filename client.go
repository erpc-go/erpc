package erpc

import (
	"context"

	"github.com/erpc-go/erpc/client"
	"github.com/erpc-go/erpc/protocol"
	"github.com/erpc-go/log"
	// "context"
	// "github.com/erpc-go/erpc/client"
)

// var DefaultClient = client.NewClient()

// func Call(ctx context.Context, addr string, req interface{}, rsp interface{}) (err error) {
// 	return DefaultClient.Call(ctx, addr, req, rsp)
// }

// Do 包级封装函
func Call(ctx context.Context, desc client.CallDesc, authinfo protocol.AuthInfo, reqBody, rspBody interface{}, opt ...map[string]string) (err error) {
	// 生成client
	c, err := client.New(desc, authinfo, reqBody, rspBody)
	if err != nil {
		log.Error(err.Error())
		return
	}
	// 用ctx执行
	return c.Do(ctx, opt...)
}

// Do 包级封装函
func Do(ctx context.Context, service string, reqBody, rspBody interface{}, opt ...map[string]string) (err error) {
	desc := client.CallDesc{
		LocalServiceName: "",
		ServiceName:      service,
		Protocol:         "test",
		Address:          "ip://127.0.0.1:8888",
	}
	// 生成client
	c, err := client.New(desc, protocol.AuthInfo{}, reqBody, rspBody)
	if err != nil {
		log.Error(err.Error())
		return
	}
	// 用ctx执行
	return c.Do(ctx, opt...)
}
