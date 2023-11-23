package client

import "bytes"

// RspBase jce空回包
type RspBase struct {
	/* 空回包，一般只判断包头返回码 */
}

// ClassName 实现gojce.Message中的接口方法
func (rb *RspBase) ClassName() string {
	return "RspBase"
}

// Encode 直接回nil
func (rb *RspBase) Encode(buf *bytes.Buffer) error {
	return nil
}

// Decode 直接回nil
func (rb *RspBase) Decode(buf *bytes.Buffer) error {
	return nil
}
