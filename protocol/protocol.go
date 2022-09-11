package protocol

import "github.com/edte/erpc/codec"

const (
	defaultMagic = 0x3bef5c
)

var (
	defaultBodyCodec codec.Codec = codec.Coder(codec.CodeTypeGob)
	defaultCodec     codec.Codec = codec.Coder(codec.CodeTypePb)
)
