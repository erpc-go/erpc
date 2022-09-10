package pb

import (
	"fmt"
	"testing"

	"github.com/edte/testpb2go/hello"
)

func TestPB(t *testing.T) {
	p := hello.HelloRequest{
		Msg: "ping",
	}
	b, err := Marshal(&p)
	pp := hello.HelloRequest{}
	fmt.Println(b, err)
	Unmarshal(b, &pp)
	fmt.Println(pp)
}
