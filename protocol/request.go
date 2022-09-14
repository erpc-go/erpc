package protocol

import (
	"io"

	"github.com/edte/erpc/codec"
	"github.com/edte/erpc/log"
)

// 请求报文格式：
// --------------------------------------------------------------
// | magic  | version |  type  | Host | Sequence | Encode | Body|
// --------------------------------------------------------------
type Request struct {
	Magic      int    // 魔数
	Version    string // rpc 版本
	Route      string // 请求 serverName.funcName
	Type       int    // 报文类型
	Sequence   int    // 序列报文号
	EncodeType string // body 编码方式
	Body       []byte // body

	encode codec.Codec // 当前的 body encode 方式
	body   interface{} // raw body type
}

func NewRequest(route string, body interface{}) *Request {
	return &Request{
		Magic:      defaultMagic,
		Version:    Version,
		Route:      route,
		Type:       int(MessageTypeRequest),
		Sequence:   getSeq(),
		EncodeType: DefaultBodyCodec.String(),
		Body:       make([]byte, 0),
		encode:     DefaultBodyCodec,
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

func (r *Request) SetHost(host string) {
	r.Route = host
}

// 编码请求
func (r *Request) EncodeTo(w io.Writer) (err error) {
	log.Debugf("begin encode request to write")

	// [step 1] 先序列化 body
	r.Body, err = r.encode.Marshal(r.body)
	if err != nil {
		log.Errorf("marshal request body to write failed, error:%s", err)
		return
	}

	log.Debugf("encode request write succ")

	// [step 2] 然后序列化整个报文
	return DefaultCodec.MarshalTo(r, w)
}

// 解码 body
func (r *Request) DecodeBody() (err error) {
	log.Debugf("begin decode request:%v", r)
	if r.encode == nil {
		log.Errorf("decode request body failed, encode nil")
		return DefaultBodyCodec.Unmarshal(r.Body, r.body)
	}
	if err = r.encode.Unmarshal(r.Body, r.body); err != nil {
		log.Errorf("decode request body failed, err:%s", err)
		return
	}
	log.Debugf("decode request body succ, body:%v", r.body)
	return
}

// 解码请求
func (r *Request) DecodeFrom(f io.Reader) (err error) {
	log.Debugf("begin decode request from reader")
	return DefaultCodec.UnmarshalFrom(f, r)
}
