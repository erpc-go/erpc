package binary

import (
	"bytes"
	"encoding/binary"
	"io"
)

// 二进制协议
type BinaryCoder struct {
}

func NewBinaryCoder() *BinaryCoder {
	return &BinaryCoder{}
}

var (
	defaultBytesOrder = binary.LittleEndian
)

func (g *BinaryCoder) Marshal(v any) ([]byte, error) {
	b := bytes.NewBuffer(make([]byte, 0))
	err := binary.Write(b, defaultBytesOrder, v)
	return b.Bytes(), err
}

func (bi *BinaryCoder) MarshalTo(v any, w io.Writer) error {
	return binary.Write(w, defaultBytesOrder, v)
}

func (g *BinaryCoder) Unmarshal(data []byte, v any) error {
	b := bytes.NewBuffer(data)
	err := binary.Read(b, defaultBytesOrder, v)
	return err
}

func (bi *BinaryCoder) UnmarshalFrom(r io.Reader, v any) error {
	return binary.Read(r, defaultBytesOrder, v)
}

func (bi *BinaryCoder) String() string {
	return "binary"
}

var (
	DefaultCoder = NewBinaryCoder()
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
