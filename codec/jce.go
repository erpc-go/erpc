package codec

import "bufio"

// TODO:
// jce 协议
type JceCoder struct {
	buf *bufio.Writer
}

func NewJceCoder() *JceCoder {
	return &JceCoder{}
}

func (g *JceCoder) Marshal(v any) ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

func (g *JceCoder) Unmarshal(data []byte, v any) error {
	panic("not implemented") // TODO: Implement
}
