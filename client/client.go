package client

import (
	"github.com/edte/erpc/codec"
	"github.com/edte/erpc/transport"
)

const (
	defaultMagic = 0x3bef5c
)

var (
	defaultCodec = codec.NewPbCoder()
)

type Client struct {
	trans       transport.Transporter
	MagicNumber int
	CodecType   codec.Codec // body codec type
}

func NewClient(opts ...Option) *Client {
	o := &clientOption{}

	c := &Client{
		trans:       nil,
		MagicNumber: defaultMagic,
		CodecType:   defaultCodec,
	}

	for _, opt := range opts {
		opt(o)
	}

	return c
}

func (c *Client) Call(addr string, req interface{}, rsp interface{}) error {
	r := NewRequest(addr, req, rsp)
	c.Do(r)
	return nil
}

func (c *Client) Do(req *Request) (Rsp *Response) {
	c.send()
	return nil
}

func (c *Client) send() {
	c.trans.Transport(nil)
}

func (c *Client) transport() transport.Transporter {
	return c.trans
}
