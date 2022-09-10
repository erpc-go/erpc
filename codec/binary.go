package codec

import (
	"bufio"
	"bytes"
	"encoding/binary"
)

// 二进制协议
type BinaryCoder struct {
	buf *bufio.Writer
}

func NewBinaryCoder() *BinaryCoder {
	return &BinaryCoder{}
}

var _ Codec = (*BinaryCoder)(nil)

var (
	defaultBytesOrder = binary.LittleEndian
)

func (g *BinaryCoder) Marshal(v any) ([]byte, error) {
	b := bytes.NewBuffer(make([]byte, 0))
	err := binary.Write(b, defaultBytesOrder, v)
	return b.Bytes(), err
}

func (g *BinaryCoder) Unmarshal(data []byte, v any) error {
	b := bytes.NewBuffer(data)
	err := binary.Read(b, defaultBytesOrder, v)
	return err
}
