package test

import (
	"github.com/erpc-go/erpc/protocol"
	"github.com/erpc-go/jce-codec"
)

type TestProtocol struct{}

func (te *TestProtocol) MarshalHeader() ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) UnmarshalHeader(_ []byte) error {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) GetCmdPattern() string {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) GetUid() uint64 {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) GetAppID() uint32 {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) GetAuthInfo() protocol.AuthInfo {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) SetAuthInfo(_ *protocol.AuthInfo) {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) GetResultCode() int32 {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) SetResultCode(_ int32) {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) GetResultMsg() string {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) SetResultMsg(_ string) {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) GetLocalServiceName() string {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) SetLocalServiceName(_ string) {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) GetServiceName() string {
	panic("not implemented") // TODO: Implement
}

// SetServiceName(string)
func (te *TestProtocol) GetProtoType() uint8 {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) SetProtoType(_ uint8) {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) GetTraceID() string {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) SetTraceID(_ string) {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) GetSpanID() uint64 {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) SetSpanID(_ uint64) {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) GetParentSpanID() uint64 {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) SetParentSpanID(_ uint64) {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) GetFlag() uint32 {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) SetFlag(_ uint32) {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) GetEnv() string {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) GetExtKv(_ string) (string, bool) {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) SetExtKv(_ string, _ string) bool {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) GetExtends() map[string]string {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) SetBodyLen(_ uint32) {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) MarshalBody(_ jce.Messager) ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) UnmarshalBody(_ []byte, _ jce.Messager) error {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) Clone() protocol.Protocol {
	panic("not implemented") // TODO: Implement
}

func (te *TestProtocol) CloneEmpty() protocol.Protocol {
	panic("not implemented") // TODO: Implement
}
