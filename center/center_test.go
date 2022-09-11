package center

import (
	"fmt"
	"testing"

	"github.com/edte/testpb2go/echo"
	"github.com/edte/testpb2go/hello"
	ping3 "github.com/edte/testpb2go/ping"
)

func show() {
	for _, s := range defaultCenter.serverList {
		fmt.Println(s)
	}
}

func TestP(t *testing.T) {
	if err := Register("demo.ping", "127.0.0.1:8080", ping3.PingRequest{}, ping3.PingResponse{}); err != nil {
		panic(err)
	}

	if err := Register("demo.ping", "127.0.0.1:134", ping3.PingRequest{}, ping3.PingResponse{}); err != nil {
		panic(err)
	}

	if err := Register("demo.ping", "127.0.0.1:9999", ping3.PingRequest{}, ping3.PingResponse{}); err != nil {
		panic(err)
	}

	if err := Register("demo.echo", "127.0.0.1:8888", echo.EchoRequest{}, echo.EchoResponse{}); err != nil {
		panic(err)
	}

	if err := Register("demo.echo", "127.0.0.1:7777", echo.EchoRequest{}, echo.EchoResponse{}); err != nil {
		panic(err)
	}

	if err := Register("hello.hello", "127.0.0.1:235", hello.HelloRequest{}, hello.HelloResponse{}); err != nil {
		panic(err)
	}

	if err := Register("hello.hello", "127.0.0.1:1111", hello.HelloRequest{}, hello.HelloResponse{}); err != nil {
		panic(err)
	}

	// show()

	fmt.Println(Discovery("demo"))
	fmt.Println(Discovery("hello"))
}
