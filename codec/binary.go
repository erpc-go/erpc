package codec

import (
	"bytes"
	"encoding/binary"
	"io"
)

// 二进制协议
type BinaryCoder struct {
	order binary.ByteOrder
}

func NewBinaryCoder() *BinaryCoder {
	return &BinaryCoder{
		order: defaultOrder,
	}
}

func (b *BinaryCoder) Marshal(v any) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	err := binary.Write(buf, b.order, v)
	return buf.Bytes(), err
}

func (b *BinaryCoder) MarshalTo(v any, w io.Writer) error {
	return binary.Write(w, b.order, v)
}

func (b *BinaryCoder) Unmarshal(data []byte, v any) error {
	buf := bytes.NewBuffer(data)
	return binary.Read(buf, b.order, v)
}

func (b *BinaryCoder) UnmarshalFrom(r io.Reader, v any) error {
	return binary.Read(r, b.order, v)
}

func (b *BinaryCoder) String() string {
	return "binary"
}
