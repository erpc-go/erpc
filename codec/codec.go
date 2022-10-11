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

type CodecType string

const (
	CodeTypeNone    CodecType = "code/none"
	CodeTypeBinary  CodecType = "code/binary"
	CodeTypeGob     CodecType = "code/gob"
	CodeTypeJce     CodecType = "code/jce"
	CodeTypeJson    CodecType = "code/json"
	CodeTypePb      CodecType = "code/pb"
	CodeTypeThrift  CodecType = "code/thrift"
	CodeTypeMsgpack CodecType = "code/msgpack"
)

func (t CodecType) String() string {
	return string(t)
}

// 默认序列化的字节序
// 默认为大端
var (
	defaultOrder = binary.LittleEndian
)

var Codecs = map[CodecType]Codec{
	CodeTypeNone:    NewRawCoder(),
	CodeTypeBinary:  NewBinaryCoder(),
	CodeTypeGob:     NewGobCoder(),
	CodeTypeJce:     NewJceCoder(),
	CodeTypeJson:    NewJsonCoder(),
	CodeTypePb:      NewPbCoder(),
	CodeTypeThrift:  NewThriftCoder(),
	CodeTypeMsgpack: NewMsgpackCoder(),
}

// 自定义注册编码方案
func RegisterCodec(t CodecType, c Codec) {
	Codecs[t] = c
}

func UnRegisterCodec(t CodecType, c Codec) {
	delete(Codecs, t)
}
