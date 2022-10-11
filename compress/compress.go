package compress

type CompressType byte

const (
	None CompressType = iota
	Gzip
	Zlib
	Huffman
)

type Compressor interface {
	Pack(data []byte) (res []byte, err error)
	UnPack(data []byte) (res []byte, err error)
}

var Compressors = map[CompressType]Compressor{
	None:    &RawCompressor{},
	Gzip:    &GzipCompressor{},
	Zlib:    &ZlipCompressor{},
	Huffman: &HuffmanCompressor{},
}

// 自定义注册编码方案
func RegisterCompressor(t CompressType, c Compressor) {
	Compressors[t] = c
}

func UnRegisterCodec(t CompressType, c Compressor) {
	delete(Compressors, t)
}
