package binary

import "github.com/edte/erpc/codec"

var (
	defaultCoder = codec.NewBinaryCoder()
)

func Marshal(v any) ([]byte, error) {
	return defaultCoder.Marshal(v)
}
func Unmarshal(data []byte, v any) error {
	return defaultCoder.Unmarshal(data, v)
}
