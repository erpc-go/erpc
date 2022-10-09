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
