package center

import (
	"fmt"
	"strings"
	"testing"
)

func TestP(t *testing.T) {
	// // defaultCenter := NewCenter()

	// if err := defaultCenter.Register("demo", "127.0.0.1:8080", []string{"heat,echo"}); err != nil {
	// 	panic(err)
	// }

	// if err := defaultCenter.Register("demo", "127.0.0.1:5533", []string{"heat,echo"}); err != nil {
	// 	panic(err)
	// }

	// if err := defaultCenter.Register("hello", "127.0.0.1:5533", []string{"hello"}); err != nil {
	// 	panic(err)
	// }

	// fmt.Printf("%v", defaultCenter.servers)

	// fmt.Println(defaultCenter.Discovery("demo"))
	// fmt.Println(defaultCenter.Discovery("hello"))
}

func TestSplit(t *testing.T) {
	s := strings.Split("demo", ".")
	fmt.Println(s)
	fmt.Println(len(s))
}
