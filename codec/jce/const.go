// Package codec implement
// 支持 jce2go 的底层库，用于基础类型的序列化
// 高级类型的序列化，由代码生成器，转换为基础类型的序列化

package jce

import (
	"unsafe"
)

// jce type
const (
	BYTE byte = iota
	SHORT
	INT
	LONG
	FLOAT
	DOUBLE
	STRING1
	STRING4
	MAP
	LIST
	StructBegin
	StructEnd
	ZeroTag
	SimpleList
)

var typeToStr = []string{
	"Byte",
	"Short",
	"Int",
	"Long",
	"Float",
	"Double",
	"String1",
	"String4",
	"Map",
	"List",
	"StructBegin",
	"StructEnd",
	"ZeroTag",
	"SimpleList",
}

func getTypeStr(t int) string {
	if t < len(typeToStr) {
		return typeToStr[t]
	}
	return "invalidType"
}

// FromInt8 NewReader(FromInt8(vec))
func FromInt8(vec []int8) []byte {
	return *(*[]byte)(unsafe.Pointer(&vec))
}
