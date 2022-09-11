package protocol

import (
	"io"
)

// 请求报文格式：
// --------------------------------------------------------------
// | magic  | version |  type  | Host | Sequence | Encode | Body|
// --------------------------------------------------------------
type Request struct {
	Magic      int    // 魔数
	Version    string // rpc 版本
	Host       string // 请求 serverName.funcName
	Type       int    // 报文类型
	Sequence   int    // 序列报文号
	EncodeType string // body 编码方式
	Body       []byte // body

	request  interface{}
	response interface{}
}

func NewRequest(server string, req interface{}, rsp interface{}) *Request {
	return &Request{
		Magic:      defaultMagic,
		Version:    Version,
		Host:       server,
		Type:       int(MessageTypeRequest),
		Sequence:   getSeq(),
		EncodeType: defaultBodyCodec.String(),
		Body:       make([]byte, 0),
		request:    req,
		response:   rsp,
	}
}

func (r *Request) Encode() (data []byte, err error) {
	// [step 1] 先序列化 body
	r.Body, err = defaultBodyCodec.Marshal(r.request)
	if err != nil {
		return nil, err
	}

	// [step 2] 然后序列化整个报文
	return defaultCodec.Marshal(r)
}

func (r *Request) EncodeTo(w io.Writer) (err error) {
	data, err := r.Encode()
	if err != nil {
		return
	}

	_, err = w.Write(data)

	return
}

func (r *Request) Decode(data []byte) (err error) {
	// [step 1] 先解码报文
	if err = defaultCodec.Unmarshal(data, r); err != nil {
		return
	}

	// [step 2] 再解码请求体
	return defaultBodyCodec.Unmarshal(r.Body, r.request)
}

func (r *Request) DecodeFrom(f io.Reader) (err error) {
	// [step 1] 先解码报文
	if err = defaultCodec.UnmarshalFrom(f, r); err != nil {
		return
	}

	// [step 2] 再解码请求体
	return defaultBodyCodec.Unmarshal(r.Body, r.request)
}

func (r *Request) SetMagic() {

}

func (r *Request) SetVersion() {

}

func (r *Request) Request() interface{} {
	return r.request
}

func (r *Request) Response() interface{} {
	return r.response
}
