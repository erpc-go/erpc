package compress

import (
	"bytes"
	"compress/gzip"
	"io"
)

// Gzip 压缩算法
type GzipCompressor struct {
}

func (c *GzipCompressor) Pack(data []byte) (res []byte, err error) {
	var b bytes.Buffer

	w := gzip.NewWriter(&b)
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

func (c *GzipCompressor) UnPack(data []byte) (res []byte, err error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return
	}

	return io.ReadAll(r)
}
