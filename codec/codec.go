package codec

// codec 接口
type Codec interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

type Marshalr interface {
	Marshal() (data []byte, err error)
}

type Unmarshaler interface {
	Unmarshal(data []byte) error
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
	coderMap[CodeTypeBinary] = NewBinaryCoder()
	coderMap[CodeTypeGob] = NewGobCoder()
	coderMap[CodeTypeJce] = NewJceCoder()
	coderMap[CodeTypeJson] = NewJsonCoder()
	coderMap[CodeTypePb] = NewPbCoder()
}

// 根据类型获取编码器
func Coder(codeType Type) Codec {
	c, ok := coderMap[codeType]
	if !ok {
		return nil
	}
	return c
}
