package protocol

type ProtocolType int

func (p ProtocolType) String() string {
	switch p {
	case ProtocolErpc:
		return "erpc"
	case ProtocolHttp:
		return "http"
	case ProtocolGrpc:
		return "grpc"
	default:
		return "unkown"
	}
}

const (
	ProtocolErpc ProtocolType = iota
	ProtocolHttp
	ProtocolGrpc
	ProtocolJson
	ProtocolTest
	ProtocolUnkown
)

func GetProtocolType(req []byte) ProtocolType {
	if req[0] == 0x2 {
		return ProtocolTest
	}

	return ProtocolUnkown
}
