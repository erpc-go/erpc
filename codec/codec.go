package codec

import (
	"io"

	"github.com/edte/erpc/codec/binary"
	"github.com/edte/erpc/codec/gob"
	"github.com/edte/erpc/codec/jce"
	"github.com/edte/erpc/codec/json"
	"github.com/edte/erpc/codec/pb"
)

// codec 接口
// TODO: 这里考虑重新设计接口, 内部加缓存？
type Codec interface {
	Marshal(v any) ([]byte, error)
	MarshalTo(v any, w io.Writer) error
	Unmarshal(data []byte, v any) error
	UnmarshalFrom(r io.Reader, v any) error
	String() string
}

type Type string

const (
	CodeTypeGob    Type = "code/gob"
	CodeTypeJce    Type = "code/jce"
	CodeTypeJson   Type = "code/json"
	CodeTypePb     Type = "code/pb"
	CodeTypeBinary Type = "code/binary"
)

func (t Type) String() string {
	return string(t)
}

var (
	coderMap map[Type]Codec = make(map[Type]Codec)
)

func init() {
	coderMap[CodeTypeBinary] = binary.DefaultCoder
	coderMap[CodeTypeGob] = gob.DefaultCoder
	coderMap[CodeTypeJce] = jce.DefaultCoder
	coderMap[CodeTypeJson] = json.DefaultCoder
	coderMap[CodeTypePb] = pb.DefaultCoder
}

// 根据类型获取编码器
func Coder(codeType Type) Codec {
	c, ok := coderMap[codeType]
	if !ok {
		return nil
	}
	return c
}
