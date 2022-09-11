package server

import "github.com/edte/erpc/codec"

type Context struct {
	header Header
	body   Body

	Request  interface{}
	Response interface{}
}

type Header struct {
	codec codec.Codec
	data  []byte
}

type Body struct {
	data  []byte
	codec codec.Codec
}

type Request struct {
}

type Response struct {
}

type Conn struct {
}

type HandlerFunc func(*Context)

var (
	handleMap = make(map[string]HandlerFunc, 200)
)

func Handle(partten string, handler HandlerFunc) {
	handleMap[partten] = handler
}

func Listen(addr string) {

}
