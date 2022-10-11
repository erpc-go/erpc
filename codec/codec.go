package codec

import (
	"encoding/binary"
	"io"
)

// codec 接口
type Codec interface {
	Marshal(v any) ([]byte, error)
	MarshalTo(v any, w io.Writer) error
	Unmarshal(data []byte, v any) error
	UnmarshalFrom(r io.Reader, v any) error
	String() string
}

var (
	defaultOrder = binary.LittleEndian
)

type Type string

const (
	CodeTypeNone    Type = "code/none"
	CodeTypeBinary  Type = "code/binary"
	CodeTypeGob     Type = "code/gob"
	CodeTypeJce     Type = "code/jce"
	CodeTypeJson    Type = "code/json"
	CodeTypePb      Type = "code/pb"
	CodeTypeThrift  Type = "code/thrift"
	CodeTypeMsgpack Type = "code/msgpack"
)

func (t Type) String() string {
	return string(t)
}

var (
	coderMap map[Type]Codec = make(map[Type]Codec)
)

func init() {
	coderMap[CodeTypeNone] = NewRawCoder()
	coderMap[CodeTypeBinary] = NewBinaryCoder()
	coderMap[CodeTypeGob] = NewGobCoder()
	coderMap[CodeTypeJce] = NewJceCoder()
	coderMap[CodeTypeJson] = NewJsonCoder()
	coderMap[CodeTypePb] = NewPbCoder()
	coderMap[CodeTypeThrift] = NewThriftCoder()
	coderMap[CodeTypeMsgpack] = NewMsgpackCoder()
}

// 根据类型获取编码器
func Coder(codeType Type) Codec {
	c, ok := coderMap[codeType]
	if !ok {
		return nil
	}
	return c
}
