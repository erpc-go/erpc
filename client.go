package erpc

import (
	"context"

	"github.com/edte/erpc/client"
)

var (
	DefaultClient = client.NewClient()
)

func Call(ctx context.Context, addr string, req interface{}, rsp interface{}) (err error) {
	return DefaultClient.Call(ctx, addr, req, rsp)
}
