package codec

import (
	"bufio"

	"github.com/golang/protobuf/proto"
)

// pb 协议
type PbCoder struct {
	buf *bufio.Writer
}

func NewPbCoder() *PbCoder {
	return &PbCoder{}
}

var _ Codec = (*PbCoder)(nil)

func (g *PbCoder) Marshal(v any) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func (g *PbCoder) Unmarshal(data []byte, v any) error {
	return proto.Unmarshal(data, v.(proto.Message))
}
