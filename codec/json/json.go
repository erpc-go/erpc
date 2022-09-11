package json

import (
	"encoding/json"
	"io"
)

// json 协议
type JsonCoder struct {
}

func NewJsonCoder() *JsonCoder {
	return &JsonCoder{}
}

func (js *JsonCoder) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (js *JsonCoder) MarshalTo(v any, w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(v)
}

func (js *JsonCoder) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (js *JsonCoder) UnmarshalFrom(r io.Reader, v any) error {
	d := json.NewDecoder(r)
	return d.Decode(r)
}

func (js *JsonCoder) String() string {
	return "jce"
}

var (
	DefaultCoder = NewJsonCoder()
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
