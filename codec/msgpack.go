package codec

import (
	"bytes"
	"io"

	"github.com/tinylib/msgp/msgp"
	"github.com/vmihailenco/msgpack/v5"
)

// msgpack 协议
type MsgpackCoder struct {
}

func NewMsgpackCoder() *MsgpackCoder {
	return &MsgpackCoder{}
}

func (m *MsgpackCoder) Marshal(v any) ([]byte, error) {
	if m, ok := v.(msgp.Marshaler); ok {
		return m.MarshalMsg(nil)
	}

	var buf bytes.Buffer

	enc := msgpack.NewEncoder(&buf)
	err := enc.Encode(v)
	return buf.Bytes(), err
}

func (m *MsgpackCoder) MarshalTo(v any, w io.Writer) (err error) {
	enc := msgpack.NewEncoder(w)
	return enc.Encode(v)
}

func (m *MsgpackCoder) Unmarshal(data []byte, v any) (err error) {
	if m, ok := v.(msgp.Unmarshaler); ok {
		_, err = m.UnmarshalMsg(data)
		return
	}

	dec := msgpack.NewDecoder(bytes.NewReader(data))
	return dec.Decode(v)
}

func (m *MsgpackCoder) UnmarshalFrom(r io.Reader, v any) (err error) {
	dec := msgpack.NewDecoder(r)
	return dec.Decode(v)
}

func (m *MsgpackCoder) String() string {
	return "Msgpack"
}
