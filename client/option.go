package client

import (
	"time"

	"github.com/edte/erpc/codec"
)

type Option func(c *clientOption)

type clientOption struct {
	MagicNumber    int
	CodecType      codec.Type
	ConnectTimeout time.Duration
	HandleTimeout  time.Duration
}
