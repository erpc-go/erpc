package protocol

import "github.com/erpc-go/erpc/codec"

const (
	defaultMagic = 0x3bef5c
)

var (
	DefaultBodyCodec codec.Codec = codec.Coder(codec.CodeTypePb)
	DefaultCodec     codec.Codec = codec.Coder(codec.CodeTypeGob)
)
