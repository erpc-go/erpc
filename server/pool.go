package server

import (
	"sync"

	"github.com/erpc-go/erpc/protocol"
	"github.com/erpc-go/erpc/protocol/test"
)

var pools = map[protocol.ProtocolType]sync.Pool{
	protocol.ProtocolErpc: {
		New: func() any {
			return nil
		},
	},
	protocol.ProtocolHttp: {
		New: func() any {
			return nil
		},
	},
	protocol.ProtocolTest: {
		New: func() any {
			return test.TestProtocol{}
		},
	},
	protocol.ProtocolGrpc: {
		New: func() any {
			return nil
		},
	},
}

func GetProtocolStruct(t protocol.ProtocolType) (p protocol.Protocol) {
	a, ok := pools[t]
	if !ok {
		return
	}
	b, ok := a.Get().(protocol.Protocol)
	if !ok {
		return
	}
	return b
}
