package pb

import (
	"fmt"
	"testing"

	"github.com/edte/testpb2go/demo"
)

func TestPB(t *testing.T) {
	p := demo.HelloRequest{
		Msg: "ping",
	}
	b, err := Marshal(&p)
	pp := demo.HelloRequest{}
	fmt.Println(b, err)
	Unmarshal(b, &pp)
	fmt.Println(pp)
}
