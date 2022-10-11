package codec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

// 裸编码，即默认不编码，只支持原生为 []byte 类型
type RawCoder struct {
	order binary.ByteOrder
}

func NewRawCoder() *RawCoder {
	return &RawCoder{
		order: defaultOrder,
	}
}

func (b *RawCoder) Marshal(v any) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}
	if data, ok := v.(*[]byte); ok {
		return *data, nil
	}

	return nil, fmt.Errorf("%T is not a []byte", v)
}

func (b *RawCoder) MarshalTo(v any, w io.Writer) (err error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	b2, err := b.Marshal(v)
	if err != nil {
		return
	}
	_, err = buf.Write(b2)
	return
}

func (b *RawCoder) Unmarshal(data []byte, v any) (err error) {
	reflect.Indirect(reflect.ValueOf(v)).SetBytes(data)
	return
}

func (b *RawCoder) UnmarshalFrom(r io.Reader, v any) (err error) {
	buf := make([]byte, 0)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return
	}
	return b.Unmarshal(buf, v)
}

func (b *RawCoder) String() string {
	return "Raw"
}
