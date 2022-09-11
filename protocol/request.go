package protocol

import (
	"io"

	"github.com/edte/erpc/codec"
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

	encode codec.Codec // 当前的 body encode 方式
	body   interface{} // raw body type
}

func NewRequest(server string, body interface{}) *Request {
	return &Request{
		Magic:      defaultMagic,
		Version:    Version,
		Host:       server,
		Type:       int(MessageTypeRequest),
		Sequence:   getSeq(),
		EncodeType: defaultBodyCodec.String(),
		Body:       make([]byte, 0),
		encode:     defaultBodyCodec,
		body:       body,
	}
}

func (r *Request) SetMagic(m int) {
	r.Magic = m
}

func (r *Request) SetEncode(c codec.Type) {
	r.encode = codec.Coder(c)
}

func (r *Request) SetBody(body interface{}) {
	r.body = body
}

// 编码请求
func (r *Request) EncodeTo(w io.Writer) (err error) {
	// [step 1] 先序列化 body
	r.Body, err = r.encode.Marshal(r.body)
	if err != nil {
		return
	}

	// [step 2] 然后序列化整个报文
	return defaultCodec.MarshalTo(r, w)
}

// 解码 body
func (r *Request) DecodeBody() (err error) {
	return r.encode.Unmarshal(r.Body, r.body)
}

// 解码请求
func (r *Request) DecodeFrom(f io.Reader) (err error) {
	return defaultCodec.UnmarshalFrom(f, r)
}
