package compress

import (
	"fmt"
	"testing"
)

func TestCompress(t *testing.T) {
	coms := []Compressor{
		&GzipCompressor{}, &HuffmanCompressor{}, &RawCompressor{}, &ZlipCompressor{},
	}

	f := func(c Compressor) {
		fmt.Println()

		s := "%5B%7B%22service%22%3A%22AttrDict%22%2C%22service_address%22%3A%22udp%40127.0.0.1%3A5353%22%7D%2C%7B%22service%22%3A%22BrasInfo%22%2C%22service_address%22%3A%22udp%40127.0.0.1%3A5353%22%7D%5D"

		t.Logf("origin len: %d", len(s))

		data, err := c.Pack([]byte(s))
		if err != nil {
			t.Fatalf("failed to pack: %v", err)
		}

		t.Logf("packed len: %d", len(data))

		s2, err := c.UnPack(data)
		if err != nil {
			t.Fatalf("failed to unpack: %v", err)
		}

		if string(s2) != s {
			t.Fatalf("unpack data is wrong")
		}
	}

	for _, c := range coms {
		f(c)
	}

}
