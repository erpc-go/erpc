package compress

import (
	"bytes"
	"compress/zlib"
	"io"
)

// zlip 压缩算法
type ZlipCompressor struct {
}

func (c *ZlipCompressor) Pack(data []byte) (res []byte, err error) {
	var b bytes.Buffer

	w := zlib.NewWriter(&b)
	if _, err = w.Write(data); err != nil {
		return
	}
	if err = w.Flush(); err != nil {
		return
	}
	if err = w.Close(); err != nil {
		return
	}

	return b.Bytes(), err
}

func (c *ZlipCompressor) UnPack(data []byte) (res []byte, err error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return
	}

	return io.ReadAll(r)
}
