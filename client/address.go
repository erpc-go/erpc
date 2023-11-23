package client

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrAddrEmpty   = errors.New("empty address")
	ErrAddrInvalid = errors.New("invalid address schema")
	ErrAddrFail    = errors.New("addrsing fail")
)

type Addressing struct {
	AddrSchema string
	Selector   Selector
	node       *Node
	beginTime  time.Time
	endTime    time.Time
}

func NewAddress(addrSchema string) (*Addressing, error) {
	if addrSchema == "" {
		return nil, ErrAddrEmpty
	}

	parts := strings.Split(addrSchema, "://")
	if len(parts) != 2 {
		return nil, ErrAddrInvalid
	}

	slt := Get(parts[0])
	node, err := slt.Select(parts[1])
	if err != nil {
		return nil, ErrAddrFail
	}

	return &Addressing{
		AddrSchema: addrSchema,
		Selector:   slt,
		node:       node,
		beginTime:  time.Now(),
	}, nil
}

func (c *Addressing) Address() string {
	if c.node == nil {
		return ""
	}
	return c.node.Addr()
}

func (c *Addressing) Cost() time.Duration {
	if c.endTime.IsZero() {
		c.endTime = time.Now()
	}
	return c.endTime.Sub(c.beginTime)
}

func (c *Addressing) Update(err error) {
	c.Selector.Update(c.node, c.Cost(), err)
}
