package protocol

import "io"

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

	req *Request
}

func NewResponse(req *Request) *Response {
	return &Response{
		Version:    Version,
		Type:       int(MessageTypeResponse),
		Sequence:   req.Sequence,
		EncodeType: defaultBodyCodec.String(),
		Body:       []byte{},
		req:        req,
	}
}

func (r *Response) Encode() (data []byte, err error) {
	// [step 1] 先序列化 body
	r.Body, err = defaultBodyCodec.Marshal(r.req.response)
	if err != nil {
		return nil, err
	}

	// [step 2] 然后序列化整个报文
	return defaultCodec.Marshal(r)
}

func (r *Response) EncodeTo(w io.Writer) (err error) {
	// [step 1] 先序列化 body
	r.Body, err = defaultBodyCodec.Marshal(r.req.response)
	if err != nil {
		return
	}

	// [step 2] 然后序列化整个报文
	return defaultCodec.MarshalTo(r, w)
}

func (r *Response) Decode(data []byte) (err error) {
	// [step 1] 先反序列化报文
	if err = defaultCodec.Unmarshal(data, r); err != nil {
		return
	}

	// [step 2] 然后反序列化 response
	if err = defaultBodyCodec.Unmarshal(r.Body, r.req.response); err != nil {
		return
	}

	return
}

func (r *Response) DecodeFrom(f io.Reader) (err error) {
	// [step 1] 先反序列化报文
	if err = defaultCodec.UnmarshalFrom(f, r); err != nil {
		return
	}

	// [step 2] 然后反序列化 response
	if err = defaultBodyCodec.Unmarshal(r.Body, r.req.response); err != nil {
		return
	}

	return
}

func (r *Response) GetStatus() int {
	return r.Status
}

func (r *Response) SetStatus(s int) {
	r.Status = s
}

func (r *Request) SetBody(data []byte) {
	r.Body = data
}
