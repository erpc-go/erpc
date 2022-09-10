package codec

import (
	"bufio"
	"bytes"
	"encoding/gob"
)

// gob 协议
type GobCoder struct {
	buf *bufio.Writer
}

func NewGobCoder() *GobCoder {
	return &GobCoder{}
}

var _ Codec = (*GobCoder)(nil)

func (g *GobCoder) Marshal(v any) ([]byte, error) {
	b := bytes.NewBuffer(make([]byte, 0))
	e := gob.NewEncoder(b)
	err := e.Encode(v)
	return b.Bytes(), err
}

func (g *GobCoder) Unmarshal(data []byte, v any) error {
	b := bytes.NewBuffer(data)
	e := gob.NewDecoder(b)
	err := e.Decode(v)
	return err
}
