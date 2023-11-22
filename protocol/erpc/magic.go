package protocol

// 魔数
// 用于 server 快速筛选那些非法请求
const (
	magicNumber byte = 0x95
)

func MagicNumber() byte {
	return magicNumber
}
