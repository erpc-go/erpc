package compress

// huffman 压缩算法
type HuffmanCompressor struct {
}

func (c *HuffmanCompressor) Pack(data []byte) ([]byte, error) {
	return data, nil
}

func (c *HuffmanCompressor) UnPack(data []byte) ([]byte, error) {
	return data, nil
}
