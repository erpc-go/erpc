package erpc

import (
	"github.com/edte/erpc/center"
)

var (
	defaultCenter = center.NewCenter()
)

func ListenCenter() {
	defaultCenter.Listen()
}
