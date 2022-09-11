package client

import "github.com/edte/erpc/register"

type Request struct {
	server string // serverName:funcName

	serverName string
	funcName   string
	addr       string // ip:port

	request  interface{}
	response interface{}

	header Header
	body   Body
}

func NewRequest(server string, req interface{}, rsp interface{}) *Request {
	return &Request{
		server:   server,
		request:  req,
		response: rsp,
		header:   Header{},
		body:     Body{},
	}
}

func (r *Request) discover() (err error) {
	addr, err := register.Discovery(r.server)
	if err != nil {
		return err
	}
	r.addr = addr
	return nil
}
