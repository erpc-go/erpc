package protocol

import (
	"io"

	"github.com/edte/erpc/codec"
)

// 响应报文格式：
// ---------------------------------------------------------
// | status  | version |  type  |  Sequence | Encode | Body|
// ---------------------------------------------------------
type Response struct {
	Status     int    // 响应状态
	Version    string // rpc 版本
	Type       int    // 报文类型
	Sequence   int    // 序列报文号
	EncodeType string // body 编码方式
	Body       []byte // body 数据

	body   interface{} // raw body type
	encode codec.Codec // body code type
}

func NewResponse() *Response {
	return &Response{
		Version:    Version,
		Type:       int(MessageTypeResponse),
		EncodeType: defaultBodyCodec.String(),
		Body:       []byte{},
		encode:     defaultBodyCodec,
	}
}

func (r *Response) SetRequence(seq int) {
	r.Sequence = seq
}

func (r *Response) SetBody(body interface{}) {
	r.body = body
}

func (r *Response) SetStatus(s int) {
	r.Status = s
}

func (r *Response) SetEncode(c codec.Type) {
	r.encode = codec.Coder(c)
}

func (r *Response) EncodeTo(w io.Writer) (err error) {
	// [step 1] 先序列化 body
	r.Body, err = r.encode.Marshal(r.body)
	if err != nil {
		return
	}

	// [step 2] 然后序列化整个报文
	return defaultCodec.MarshalTo(r, w)
}

func (r *Response) DecodeFrom(f io.Reader) (err error) {
	// [step 1] 先反序列化报文
	if err = defaultCodec.UnmarshalFrom(f, r); err != nil {
		return
	}

	// [step 2] 然后反序列化 body
	if err = r.encode.Unmarshal(r.Body, r.body); err != nil {
		return
	}

	return
}
