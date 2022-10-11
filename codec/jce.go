package codec

import (
	"io"

	jce2 "github.com/erpc-go/jce-codec"
)

// jce 协议
type JceCoder struct {
}

func NewJceCoder() *JceCoder {
	return &JceCoder{}
}

func (jc *JceCoder) Marshal(v any) ([]byte, error) {
	return jce2.Marshal(v)
}

func (jc *JceCoder) MarshalTo(v any, w io.Writer) error {
	return jce2.MarshalTo(v, w)
}

func (jc *JceCoder) Unmarshal(data []byte, v any) error {
	return jce2.Unmarshal(data, v)
}

func (jc *JceCoder) UnmarshalFrom(r io.Reader, v any) error {
	return jce2.UnmarshalFrom(r, v)
}

func (jc *JceCoder) String() string {
	return "jce"
}
