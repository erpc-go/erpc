package protocol

import (
	"fmt"
	"testing"

	"github.com/edte/testpb2go/hello"
)

func TestNewResponse(t *testing.T) {
	r := NewResponse(NewRequest("hello.ping", &hello.HelloRequest{}, &hello.HelloResponse{}))
	data, err := r.Encode()
	fmt.Println(data, err)

	rr := Response{
		req: NewRequest("hello.ping", &hello.HelloRequest{}, &hello.HelloResponse{}),
	}

	fmt.Println()
	fmt.Println(rr.Decode(data))
}
