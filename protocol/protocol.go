package protocol

import "github.com/erpc-go/jce-codec"

type Protocol interface {
	Header
	Body
	Clone() Protocol
	CloneEmpty() Protocol
}

// TME协议通用接口
type Header interface {
	MarshalHeader() ([]byte, error)
	UnmarshalHeader([]byte) error
	GetCmdPattern() string
	GetUid() uint64
	GetAppID() uint32
	GetAuthInfo() AuthInfo
	SetAuthInfo(*AuthInfo)
	GetResultCode() int32
	SetResultCode(int32)
	GetResultMsg() string
	SetResultMsg(string)
	GetLocalServiceName() string
	SetLocalServiceName(string)
	GetServiceName() string
	// SetServiceName(string)
	GetProtoType() uint8
	SetProtoType(uint8)
	GetTraceID() string
	SetTraceID(string)
	GetSpanID() uint64
	SetSpanID(uint64)
	GetParentSpanID() uint64
	SetParentSpanID(uint64)
	GetFlag() uint32
	SetFlag(uint32)
	GetEnv() string
	GetExtKv(string) (string, bool)
	SetExtKv(string, string) bool
	GetExtends() map[string]string
	SetBodyLen(uint32)
}

type Body interface {
	MarshalBody(jce.Messager) ([]byte, error)
	UnmarshalBody([]byte, jce.Messager) error
}
