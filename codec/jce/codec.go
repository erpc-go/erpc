package jce

import (
	"bytes"
	"errors"
	"io"
)

type Messager interface {
	ResetDefault()
	ReadFrom(r io.Reader) (err error)
	ReadBlock(r io.Reader, tag byte, require bool) error
	WriteTo(w io.Writer) (err error)
	WriteBlock(w io.Writer, tag byte) (err error)
}

func Marshal(v any) (data []byte, err error) {
	m, ok := v.(Messager)
	if !ok {
		return nil, errors.New("not jce Messager type")
	}
	b := bytes.NewBuffer(make([]byte, 0))
	err = m.WriteTo(b)
	return b.Bytes(), err
}

func MarshalTo(v any, w io.Writer) error {
	m, ok := v.(Messager)
	if !ok {
		return errors.New("not jce Messager type")
	}
	return m.WriteTo(w)
}

func Unmarshal(data []byte, v any) error {
	m, ok := v.(Messager)
	if !ok {
		return errors.New("not jce Messager type")
	}
	b := bytes.NewBuffer(data)
	return m.ReadFrom(b)
}

func UnmarshalFrom(r io.Reader, v any) error {
	m, ok := v.(Messager)
	if !ok {
		return errors.New("not jce Messager type")
	}
	return m.ReadFrom(r)
}
