package erpc

import (
	"context"

	"github.com/erpc-go/erpc/client"
	"github.com/erpc-go/erpc/protocol"
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
	c := client.New(desc, authinfo, reqBody, rspBody)
	// 用ctx执行
	return c.Do(ctx, opt...)
}
