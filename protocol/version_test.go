package protocol_test

import (
	"fmt"
	"testing"

	"github.com/edte/erpc/protocol"
)

func TestVersion(t *testing.T) {
	a := ([]byte)(protocol.Version)
	fmt.Println(len(a))
}
