package jce

import (
	"io"
)

// TODO:
// jce 协议
type JceCoder struct {
}

func NewJceCoder() *JceCoder {
	return &JceCoder{}
}

func (jc *JceCoder) Marshal(v any) ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

func (jc *JceCoder) MarshalTo(v any, w io.Writer) error {
	panic("not implemented") // TODO: Implement
}

func (jc *JceCoder) Unmarshal(data []byte, v any) error {
	panic("not implemented") // TODO: Implement
}

func (jc *JceCoder) UnmarshalFrom(r io.Reader, v any) error {
	panic("not implemented") // TODO: Implement
}

func (jc *JceCoder) String() string {
	return "jce"
}

var (
	DefaultCoder = NewJceCoder()
)

func Marshal(v any) ([]byte, error) {
	return DefaultCoder.Marshal(v)
}

func MarshalTo(v any, w io.Writer) error {
	return DefaultCoder.MarshalTo(v, w)
}

func Unmarshal(data []byte, v any) error {
	return DefaultCoder.Unmarshal(data, v)
}

func UnmarshalFrom(r io.Reader, v any) error {
	return DefaultCoder.UnmarshalFrom(r, v)

}
