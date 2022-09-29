package protocol

import (
	"io"

	"github.com/erpc-go/erpc/codec"
	"github.com/erpc-go/erpc/log"
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
		EncodeType: DefaultBodyCodec.String(),
		Body:       []byte{},
		encode:     DefaultBodyCodec,
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

func (r *Response) SetEncode(c codec.Codec) {
	r.encode = c
}

func (r *Response) EncodeTo(w io.Writer) (err error) {
	log.Debugf("begin encode response to write")

	// [step 1]  如果 body 为空，则直接序列化报文
	if r.body == nil {
		log.Errorf("marshal response body to write failed, error:%s", err)
		return DefaultCodec.MarshalTo(r, w)
	}

	log.Debugf("begin encode response body to write succ")

	// [step 2] 先序列化 body
	r.Body, err = r.encode.Marshal(r.body)
	if err != nil {
		log.Errorf("marshal response body to write failed, error:%s", err)
		return
	}

	// [step 3] 然后序列化整个报文
	return DefaultCodec.MarshalTo(r, w)
}

func (r *Response) DecodeFrom(f io.Reader) (err error) {
	log.Debugf("begin decode response from reader")

	// [step 1] 先反序列化报文
	if err = DefaultCodec.UnmarshalFrom(f, r); err != nil {
		log.Errorf("unmarshal response form reader failed, error:%s", err)
		return
	}

	// [step 2] 然后反序列化 body
	if err = r.encode.Unmarshal(r.Body, r.body); err != nil {
		log.Errorf("unmarshal response body form reader failed, error:%s", err)
		return
	}

	log.Debugf("decode response from reader succ")

	return
}
