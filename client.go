package erpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/edte/erpc/client"
)

var (
	DefaultClient = client.NewClient()
)

func Call(ctx context.Context, addr string, req interface{}, rsp interface{}) (err error) {
	res := &client.CallRes{
		Done: make(chan struct{}),
		Err:  err,
	}

	go func() {
		DefaultClient.Call(addr, req, rsp, res)
	}()

	select {
	case <-ctx.Done():
		return errors.New(fmt.Sprintf("call %s timeout", addr))
	case <-res.Done:
		return res.Err
	}
}
