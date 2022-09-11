package pb

import (
	"io"

	"github.com/golang/protobuf/proto"
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

func (pb *PbCoder) MarshalTo(v any, w io.Writer) error {
	panic("not implemented") // TODO: Implement
}

func (pb *PbCoder) Unmarshal(data []byte, v any) error {
	return proto.Unmarshal(data, v.(proto.Message))
}

func (pb *PbCoder) UnmarshalFrom(r io.Reader, v any) error {
	panic("not implemented") // TODO: Implement
}

func (pb *PbCoder) String() string {
	return "pb"
}

var (
	DefaultCoder = NewPbCoder()
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
