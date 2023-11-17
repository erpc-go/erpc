package log_test

import (
	"testing"

	"github.com/erpc-go/erpc/log"
)

func TestSetLogLevel(t *testing.T) {
	log.Debug("debug")
	log.Error("error")
}
