package protocol

import (
	"fmt"
	"testing"

	"github.com/edte/testpb2go/echo"
)

func TestNewRequest(t *testing.T) {
	r := NewRequest("hello.echo", &echo.EchoRequest{}, &echo.EchoResponse{})
	data, err := r.Encode()
	fmt.Println(data, err)

	rr := NewRequest("he.ho", &echo.EchoRequest{}, &echo.EchoResponse{})
	err = rr.Decode(data)
	fmt.Println(err)

	fmt.Println(rr)
}
