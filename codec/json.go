package codec

import (
	"bufio"
	"encoding/json"
)

// json 协议
type JsonCoder struct {
	buf *bufio.Writer
}

func NewJsonCoder() *JceCoder {
	return &JceCoder{}
}

var _ Codec = (*JsonCoder)(nil)

func (g *JsonCoder) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (g *JsonCoder) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
