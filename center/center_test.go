package center

import (
	"fmt"
	"testing"
)

func show() {
	for _, s := range defaultCenter.serverList {
		fmt.Println(s)
	}
}

func TestP(t *testing.T) {
	if err := defaultCenter.register("demo", "127.0.0.1:8080", []string{"ping,echo"}); err != nil {
		panic(err)
	}

	if err := defaultCenter.register("demo", "127.0.0.1:5533", []string{"ping,echo"}); err != nil {
		panic(err)
	}

	if err := defaultCenter.register("hello", "127.0.0.1:5533", []string{"hello"}); err != nil {
		panic(err)
	}

	show()

	fmt.Println(defaultCenter.discovery("demo"))
	fmt.Println(defaultCenter.discovery("hello"))
}
