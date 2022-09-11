package erpc

import "github.com/edte/erpc/client"

var (
	DefaultClient = client.NewClient()
)

func Call(addr string, req interface{}, rsp interface{}) (err error) {
	return DefaultClient.Call(addr, req, rsp)
}
