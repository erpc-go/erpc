package gob

import (
	"bytes"
	"encoding/gob"
	"io"
)

// gob 协议
type GobCoder struct {
}

func NewGobCoder() *GobCoder {
	return &GobCoder{}
}

func (g *GobCoder) Marshal(v any) ([]byte, error) {
	b := bytes.NewBuffer(make([]byte, 0))
	e := gob.NewEncoder(b)
	err := e.Encode(v)
	return b.Bytes(), err
}

func (g *GobCoder) MarshalTo(v any, w io.Writer) error {
	e := gob.NewEncoder(w)
	return e.Encode(v)
}

func (g *GobCoder) Unmarshal(data []byte, v any) error {
	b := bytes.NewBuffer(data)
	e := gob.NewDecoder(b)
	err := e.Decode(v)
	return err
}

func (g *GobCoder) UnmarshalFrom(r io.Reader, v any) error {
	d := gob.NewDecoder(r)
	return d.Decode(v)
}

func (g *GobCoder) String() string {
	return "gob"
}

var (
	DefaultCoder = NewGobCoder()
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
