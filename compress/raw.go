package compress

// 不处理
type RawCompressor struct {
}

func (c *RawCompressor) Pack(data []byte) ([]byte, error) {
	return data, nil
}

func (c *RawCompressor) UnPack(data []byte) ([]byte, error) {
	return data, nil
}
