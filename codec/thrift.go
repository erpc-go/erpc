package codec

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/apache/thrift/lib/go/thrift"
)

// thrift 协议
type ThriftCoder struct {
}

func NewThriftCoder() *ThriftCoder {
	return &ThriftCoder{}
}

func (t *ThriftCoder) Marshal(v any) ([]byte, error) {
	b := thrift.NewTMemoryBufferLen(1024)
	p := thrift.NewTBinaryProtocolFactoryConf(&thrift.TConfiguration{}).GetProtocol(b)

	tt := &thrift.TSerializer{
		Transport: b,
		Protocol:  p,
	}

	tt.Transport.Close()

	if msg, ok := v.(thrift.TStruct); ok {
		return tt.Write(context.Background(), msg)
	}

	return nil, errors.New("type assertion failed")
}

func (t *ThriftCoder) MarshalTo(v any, w io.Writer) (err error) {
	b, err := t.Marshal(v)
	if err != nil {
		return
	}
	_, err = w.Write(b)
	return
}

func (t *ThriftCoder) Unmarshal(data []byte, v any) error {
	tt := thrift.NewTMemoryBufferLen(1024)
	p := thrift.NewTBinaryProtocolFactoryConf(&thrift.TConfiguration{}).GetProtocol(tt)

	d := &thrift.TDeserializer{
		Transport: tt,
		Protocol:  p,
	}
	d.Transport.Close()

	return d.Read(context.Background(), v.(thrift.TStruct), data)
}

func (t *ThriftCoder) UnmarshalFrom(r io.Reader, v any) (err error) {
	b := bytes.NewBuffer(make([]byte, 0))
	_, err = b.ReadFrom(r)
	if err != nil {
		return
	}

	return t.Unmarshal(b.Bytes(), v)
}

func (Thrift *ThriftCoder) String() string {
	return "Thrift"
}
