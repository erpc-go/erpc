package protocol

import "github.com/erpc-go/erpc/codec"

const (
	defaultMagic = 0x3bef5c
)

var (
	DefaultBodyCodec codec.Codec = codec.Codecs[codec.CodeTypePb]
	DefaultCodec     codec.Codec = codec.Codecs[codec.CodeTypeGob]
)
