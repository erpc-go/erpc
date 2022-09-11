package client

type Response struct {
	header Header
	body   Body
}

func NewResponse() *Response {
	return &Response{}
}
