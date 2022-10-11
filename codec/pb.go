package codec

import (
	"bytes"
	"io"

	"github.com/gogo/protobuf/proto"
)

// pb 协议
type PbCoder struct {
}

func NewPbCoder() *PbCoder {
	return &PbCoder{}
}

func (pb *PbCoder) Marshal(v any) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func (pb *PbCoder) MarshalTo(v any, w io.Writer) (err error) {
	b, err := proto.Marshal(v.(proto.Message))
	if err != nil {
		return
	}
	_, err = w.Write(b)
	return
}

func (pb *PbCoder) Unmarshal(data []byte, v any) error {
	return proto.Unmarshal(data, v.(proto.Message))
}

func (pb *PbCoder) UnmarshalFrom(r io.Reader, v any) (err error) {
	b := bytes.NewBuffer(make([]byte, 0))
	if _, err = b.ReadFrom(r); err != nil {
		return
	}

	return proto.Unmarshal(b.Bytes(), v.(proto.Message))
}

func (pb *PbCoder) String() string {
	return "pb"
}
