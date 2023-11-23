package http

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
)

// HTTP 1.x 扩展Header
// 命名遵循HTTP首部规范: 驼峰+'-'字符分隔: https://cs.opensource.google/go/go/+/refs/tags/go1.17.1:src/net/textproto/reader.go;drc=refs%2Ftags%2Fgo1.17.1;bpv=0;bpt=0;l=635
// golang源码自动将Header统一转成标准格式
const (
	ServiceName  = "Service-Name"
	TraceID      = "Trace-Id"
	SpanID       = "Span-Id"
	ParentSpanID = "Parent-Span-Id"
	Flag         = "Flag"
	Env          = "Env"
	// 以下为鉴权相关
	Uid       = "Uid"
	TokenType = "Token-Type"
	AuthType  = "Auth-Type"
	OpenID    = "Open-Id"
	Ticket    = "Ticket"
	Appid     = "App-Id"
	OpenAppID = "Open-App-Id"
)

// HTTP 1.x协议(暂不支持Chunked)
// 以下所有的API, Getter操作的都是HTTP Request, Setter操作的都是HTTP Response
// HTTPHeader HTTP 1.x协议首部
type HTTPHeader struct {
	request *http.Request
	// response (net/http没有将写Response作为可导出方案)
	responseCode          int32             // 返回码
	responseMsg           string            // 提示语
	responseContentLength int64             // HTTP响应Body长度
	responseProtoType     uint8             // HTTP响应Body MIME类型
	extMap                map[string]string // k-v结构的协议扩展首部
}

// NewHTTPHeader() 创建新的HTTP首部结构体
func NewHTTPHeader() *HTTPHeader {
	return &HTTPHeader{
		extMap: make(map[string]string),
	}
}

// Marshal 构造HTTP首部
// golang官方未提供形如: http.WriteResponse的库, 自己实现
func (h *HTTPHeader) Marshal() ([]byte, error) {
	if h == nil || h.request == nil {
		return nil, nil
	}
	// 状态栏(框架自动生成)
	statusLine := fmt.Sprintf("HTTP/%d.%d 200 OK\r\n", h.request.ProtoMajor, h.request.ProtoMinor)
	// 首部
	var header string
	// Date(预留)
	header += fmt.Sprintf("Date: %s\r\n", time.Now().UTC().Format(http.TimeFormat))
	// Content-Length(预留)
	header += fmt.Sprintf("Content-Length: %d\r\n", h.GetResponseContentLength())
	// Content-Type(预留, 但业务可通过API设置)
	contentType := h.responseProtoType
	if contentType != TmeProtoTypeJce ||
		contentType != TmeProtoTypePb ||
		contentType != TmeProtoTypeJSON {
		contentType = h.GetProtoType()
	}
	switch contentType {
	case TmeProtoTypeJce:
		header += fmt.Sprintf("Content-Type: application/jce\r\n")
	case TmeProtoTypePb:
		header += fmt.Sprintf("Content-Type: application/pb\r\n")
	case TmeProtoTypeJSON:
		header += fmt.Sprintf("Content-Type: application/json\r\n")
	default:
		header += fmt.Sprintf("Content-Type: application/x-www-form-urlencoded\r\n")
	}
	// 自定义
	if h.extMap != nil {
		for k, v := range h.extMap {
			if k == "Date" || k == "Content-Length" || k == "Content-Type" {
				continue
			}
			header += fmt.Sprintf("%s: %s\r\n", k, v)
		}
	}
	// CRLF
	crlf := fmt.Sprintf("\r\n")
	return []byte(statusLine + header + crlf), nil
}

// Unmarshal 解析HTTP首部
// golang官方提供http.ReadRequest库
func (h *HTTPHeader) Unmarshal(data []byte) error {
	var err error
	h.request, err = http.ReadRequest(bufio.NewReader(bytes.NewReader(data)))
	if err != nil {
		return err
	}
	return nil
}

// GetCmdPattern 构造CmdPattern
func (h *HTTPHeader) GetCmdPattern() string {
	if h == nil || h.request == nil || h.request.URL == nil {
		return ""
	}
	return h.request.URL.Path
}

// GetUid 获取Uid
func (h *HTTPHeader) GetUid() uint64 {
	if h == nil || h.request == nil || h.request.Header == nil {
		return 0
	}
	if len(h.request.Header[Uid]) == 0 {
		return 0
	}
	id, _ := strconv.Atoi(h.request.Header[Uid][0])
	return uint64(id)
}

// SetUid 设置Uid
func (h *HTTPHeader) SetUid(uid uint64) {
	if h == nil || h.extMap == nil {
		return
	}
	h.SetExtKv(Uid, strconv.FormatUint(uid, 10))
}

// GetTokenType 获取TokenType
func (h *HTTPHeader) GetTokenType() uint32 {
	if h == nil || h.request == nil || h.request.Header == nil {
		return 0
	}
	if len(h.request.Header[TokenType]) == 0 {
		return 0
	}
	id, _ := strconv.Atoi(h.request.Header[TokenType][0])
	return uint32(id)
}

// SetTokenType 设置TokenType
func (h *HTTPHeader) SetTokenType(v uint32) {
	if h == nil || h.extMap == nil {
		return
	}
	h.SetExtKv(TokenType, strconv.FormatUint(uint64(v), 10))
}

// GetAuthType 获取AuthType
func (h *HTTPHeader) GetAuthType() uint32 {
	if h == nil || h.request == nil || h.request.Header == nil {
		return 0
	}
	if len(h.request.Header[AuthType]) == 0 {
		return 0
	}
	id, _ := strconv.Atoi(h.request.Header[AuthType][0])
	return uint32(id)
}

// SetAuthType 设置AuthType
func (h *HTTPHeader) SetAuthType(v uint32) {
	if h == nil || h.extMap == nil {
		return
	}
	h.SetExtKv(AuthType, strconv.FormatUint(uint64(v), 10))
}

// GetOpenID 获取OpenID
func (h *HTTPHeader) GetOpenID() string {
	if h == nil || h.request == nil || h.request.Header == nil {
		return ""
	}
	if len(h.request.Header[OpenID]) == 0 {
		return ""
	}
	return h.request.Header[OpenID][0]
}

// SetOpenID 设置OpenID
func (h *HTTPHeader) SetOpenID(v string) {
	if h == nil || h.extMap == nil {
		return
	}
	h.SetExtKv(OpenID, v)
}

// GetTicket 获取Ticket
func (h *HTTPHeader) GetTicket() string {
	if h == nil || h.request == nil || h.request.Header == nil {
		return ""
	}
	if len(h.request.Header[Ticket]) == 0 {
		return ""
	}
	return h.request.Header[Ticket][0]
}

// SetTicket 设置Ticket
func (h *HTTPHeader) SetTicket(v string) {
	if h == nil || h.extMap == nil {
		return
	}
	h.SetExtKv(Ticket, v)
}

// GetClientIP 获取客户端IP
func (h *HTTPHeader) GetClientIP() uint32 {
	if h == nil || h.request == nil || h.request.Header == nil {
		return 0
	}
	if len(h.request.Header["X-Forwarded-For"]) == 0 {
		return 0
	}
	clientIP := net.ParseIP(h.request.Header["X-Forwarded-For"][0])
	if len(clientIP) == 0 {
		return 0
	}
	return binary.BigEndian.Uint32(clientIP)
}

// SetClientIP 设置ClientIP
func (h *HTTPHeader) SetClientIP(v uint32) {
	if h == nil || h.extMap == nil {
		return
	}
	h.SetExtKv("X-Forwarded-For", strconv.FormatUint(uint64(v), 10))
}

// GetAppID 获取App ID
func (h *HTTPHeader) GetAppID() uint32 {
	if h == nil || h.request == nil || h.request.Header == nil {
		return 0
	}
	if len(h.request.Header[Appid]) == 0 {
		return 0
	}
	id, _ := strconv.Atoi(h.request.Header[Appid][0])
	return uint32(id)
}

// SetAppID 设置AppID
func (h *HTTPHeader) SetAppID(v uint32) {
	if h == nil || h.extMap == nil {
		return
	}
	h.SetExtKv(Appid, strconv.FormatUint(uint64(v), 10))
}

// GetOpenAppID 获取OpenAppID
func (h *HTTPHeader) GetOpenAppID() string {
	if h == nil || h.request == nil || h.request.Header == nil {
		return ""
	}
	if len(h.request.Header[OpenAppID]) == 0 {
		return ""
	}
	return h.request.Header[OpenAppID][0]
}

// SetOpenAppID 设置OpenAppID
func (h *HTTPHeader) SetOpenAppID(v string) {
	if h == nil || h.extMap == nil {
		return
	}
	h.SetExtKv(OpenAppID, v)
}

// GetAuthInfo 获取AuthInfo
func (h *HTTPHeader) GetAuthInfo() AuthInfo {
	if h == nil || h.request == nil || h.request.Header == nil {
		return AuthInfo{}
	}
	return AuthInfo{
		UID:       h.GetUid(),
		TokenType: h.GetTokenType(),
		AuthType:  h.GetAuthType(),
		OpenID:    h.GetOpenID(),
		Ticket:    h.GetTicket(),
		ClientIP:  h.GetClientIP(),
		AppID:     h.GetAppID(),
		OpenAppID: h.GetOpenAppID(),
	}
}

// SetAuthInfo 设置AuthInfo
func (h *HTTPHeader) SetAuthInfo(v *AuthInfo) {
	if h == nil || h.extMap == nil {
		return
	}
	h.SetUid(v.UID)
	h.SetTokenType(v.TokenType)
	h.SetAuthType(v.AuthType)
	h.SetOpenID(v.OpenID)
	h.SetTicket(v.Ticket)
	h.SetClientIP(v.ClientIP)
	h.SetAppID(v.AppID)
	h.SetOpenAppID(v.OpenAppID)
}

// GetResultCode 获取业务错误码
func (h *HTTPHeader) GetResultCode() int32 {
	if h == nil {
		return 0
	}
	return h.responseCode
}

// SetResultCode 设置业务错误码
func (h *HTTPHeader) SetResultCode(code int32) {
	if h == nil {
		return
	}
	h.responseCode = code
}

// GetResultMsg 获取业务错误信息
func (h *HTTPHeader) GetResultMsg() string {
	if h == nil {
		return ""
	}
	return h.responseMsg
}

// SetResultMsg 设置错误信息
func (h *HTTPHeader) SetResultMsg(msg string) {
	if h == nil {
		return
	}
	h.responseMsg = msg
}

// GetResponseContentLength 获取HTTP相应Content-Length
func (h *HTTPHeader) GetResponseContentLength() int64 {
	if h == nil {
		return 0
	}
	return h.responseContentLength
}

// SetResponseContentLength 设置HTTP相应Content-Length
func (h *HTTPHeader) SetResponseContentLength(length int64) {
	if h == nil {
		return
	}
	h.responseContentLength = length
}

// GetLocalServiceName 获取Remote Service Name
func (h *HTTPHeader) GetLocalServiceName() string {
	return ""
}

// SetLocalServiceName 设置Remote Service Name
func (h *HTTPHeader) SetLocalServiceName(string) {
}

// GetServiceName 获取Local Service Name
func (h *HTTPHeader) GetServiceName() string {
	if h == nil || h.request == nil || h.request.Header == nil {
		return ""
	}
	if len(h.request.Header[ServiceName]) == 0 {
		return ""
	}
	return h.request.Header[ServiceName][0]
}

// GetProtoType 获取body MIME类型
func (h *HTTPHeader) GetProtoType() uint8 {
	if h == nil || h.request == nil || h.request.Header == nil {
		return TmeProtoTypeUnknown
	}
	if len(h.request.Header["Content-Type"]) == 0 {
		return TmeProtoTypeUnknown
	}
	cts := strings.Split(h.request.Header["Content-Type"][0], ";")
	switch cts[0] {
	case "application/json":
		return TmeProtoTypeJSON
	case "application/jce":
		return TmeProtoTypeJce
	case "application/pb":
		return TmeProtoTypePb
	default:
		return TmeProtoTypeUnknown
	}
}

// SetProtoType 设置body MIME类型
func (h *HTTPHeader) SetProtoType(v uint8) {
	if h == nil {
		return
	}
	h.responseProtoType = v
}

// GetTraceID 获取TraceID
func (h *HTTPHeader) GetTraceID() string {
	if h == nil || h.request == nil || h.request.Header == nil {
		return ""
	}
	if len(h.request.Header[TraceID]) == 0 {
		return ""
	}
	return h.request.Header[TraceID][0]
}

// SetTraceID 设置TraceID
func (h *HTTPHeader) SetTraceID(v string) {
	if h == nil || h.extMap == nil {
		return
	}
	h.SetExtKv(TraceID, v)
}

// GetSpanID 获取Span ID
func (h *HTTPHeader) GetSpanID() uint64 {
	if h == nil || h.request == nil || h.request.Header == nil {
		return 0
	}
	if len(h.request.Header[SpanID]) == 0 {
		return 0
	}
	id, _ := strconv.Atoi(h.request.Header[SpanID][0])
	return uint64(id)
}

// SetSpanID 设置Span ID
func (h *HTTPHeader) SetSpanID(v uint64) {
	if h == nil || h.extMap == nil {
		return
	}
	h.SetExtKv(SpanID, strconv.FormatUint(v, 10))
}

// GetParentSpanID 获取ParentSpanID
func (h *HTTPHeader) GetParentSpanID() uint64 {
	if h == nil || h.request == nil || h.request.Header == nil {
		return 0
	}
	if len(h.request.Header[ParentSpanID]) == 0 {
		return 0
	}
	id, _ := strconv.Atoi(h.request.Header[ParentSpanID][0])
	return uint64(id)
}

// SetParentSpanID 设置ParentSpanID
func (h *HTTPHeader) SetParentSpanID(v uint64) {
	if h == nil || h.extMap == nil {
		return
	}
	h.SetExtKv(ParentSpanID, strconv.FormatUint(v, 10))
}

// GetFlag 获取Flag
func (h *HTTPHeader) GetFlag() uint32 {
	if h == nil || h.request == nil || h.request.Header == nil {
		return 0
	}
	if len(h.request.Header[Flag]) == 0 {
		return 0
	}
	id, _ := strconv.Atoi(h.request.Header[Flag][0])
	return uint32(id)
}

// SetFlag 设置Flag
func (h *HTTPHeader) SetFlag(v uint32) {
	if h == nil || h.extMap == nil {
		return
	}
	h.SetExtKv(Flag, strconv.FormatUint(uint64(v), 10))
}

// GetEnv 获取环境标识
func (h *HTTPHeader) GetEnv() string {
	if h == nil || h.request == nil || h.request.Header == nil {
		return ""
	}
	if len(h.request.Header[Env]) == 0 {
		return ""
	}
	return h.request.Header[Env][0]
}

// GetExtKv 获取HTTP Header
func (h *HTTPHeader) GetExtKv(k string) (string, bool) {
	if h == nil || h.request == nil || h.request.Header == nil {
		return "", false
	}
	if len(h.request.Header[k]) == 0 {
		return "", false
	}
	var v string = h.request.Header[k][0]
	for i := 1; i < len(h.request.Header[k]); i++ {
		v += ", " + h.request.Header[k][i]
	}
	return v, false
}

// SetExtKv 设置HTTP Header
func (h *HTTPHeader) SetExtKv(k string, v string) bool {
	if h == nil || h.extMap == nil {
		return false
	}
	h.extMap[k] = v
	return false
}

// GetExtends 获取所有扩展首部
func (h *HTTPHeader) GetExtends() map[string]string {
	if h == nil {
		return nil
	}
	return h.extMap
}

// Clone HTTP首部深复制
func (h *HTTPHeader) Clone() TMEHead {
	newer := NewHTTPHeader()
	// TODO
	return newer
}

// HTTPCheck HTTP 1.x协议包完整性校验(暂不支持Chunked)
func HTTPCheck(data []byte) (int, error) {
	var index int = 0

	// 解析请求行
	for index < len(data) && data[index] != '\n' {
		index++
	}
	if index == len(data) {
		return 0, nil
	}
	if index <= 1 {
		return 0, fmt.Errorf("invalid request line: %v", data[0:index])
	}
	if data[index-1] != '\r' {
		return 0, fmt.Errorf("invalid request line: %v", data[0:index])
	}
	index++

	var headerLength, contentLength int = index, 0
	// 解析请求首部
	var end int = index
	for ; end < len(data); end++ {
		if data[end] == '\n' && end > index && data[end-1] == '\r' {
			k, v := mimeHeader(data[index:end])
			if strings.ToLower(string(k)) == strings.ToLower("Content-Length") {
				length, err := strconv.Atoi(string(v))
				if err != nil {
					cat.Logf("Content-Length Atoi failed, %s\n", err)
					return 0, fmt.Errorf("Content-Length Atoi failed, %s", err)
				}
				contentLength = length
			}
			if end == index+1 {
				headerLength = end + 1
				break
			}
			index = end + 1
			end = index
		}
	}
	if end == len(data) {
		return 0, nil
	}

	// 包长度
	if contentLength > len(data[headerLength:]) {
		return 0, nil
	}

	return headerLength + contentLength, nil
}

// HTTPBodyMarshal 序列化包体
func HTTPBodyMarshal(head *HTTPHeader, rsp interface{}) ([]byte, error) {
	if head == nil {
		return nil, fmt.Errorf("HTTPHeader is nil")
	}
	// Content-Type(预留, 但业务可通过API设置)
	contentType := head.responseProtoType
	if contentType != TmeProtoTypeJce ||
		contentType != TmeProtoTypePb ||
		contentType != TmeProtoTypeJSON {
		contentType = head.GetProtoType()
	}
	var err error
	var bodyBuf []byte
	switch contentType {
	case TmeProtoTypeJSON:
		bodyBuf, err = json.Marshal(rsp)
		if err != nil {
			return nil, err
		}
	case TmeProtoTypeJce:
		rsp, _ := rsp.(gojce.Message)
		var b bytes.Buffer // TODO: 感觉这里可以预留 header 的长度 后面就不用复制了
		err = rsp.Encode(&b)
		if err != nil {
			return nil, err
		}
		bodyBuf = b.Bytes()
	case TmeProtoTypePb:
		rsp, _ := rsp.(proto.Message)
		bodyBuf, err = proto.Marshal(rsp)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid Content-Type:%d", contentType)
	}

	return bodyBuf, nil
}

// HTTPBodyUnmarshal 反序列化包体
func HTTPBodyUnmarshal(head *HTTPHeader, data []byte, req interface{}) error {
	var err error
	if head == nil {
		head = NewHTTPHeader()
	}
	if head.request == nil {
		err = head.Unmarshal(data)
		if err != nil {
			return err
		}
	}
	// 读取HTTP Body
	defer head.request.Body.Close()
	body, err := ioutil.ReadAll(head.request.Body)
	if err != nil {
		return err
	}

	// Body协议解析
	// https://cs.opensource.google/go/go/+/refs/tags/go1.17.1:src/net/textproto/reader.go;drc=refs%2Ftags%2Fgo1.17.1;bpv=0;bpt=0;l=635
	// golang源码自动将Header统一转成标准格式
	if len(head.request.Header) == 0 || len(head.request.Header["Content-Type"]) == 0 {
		return fmt.Errorf("Content-Type empty")
	}
	cts := strings.Split(head.request.Header["Content-Type"][0], ";")
	switch cts[0] {
	case "application/json":
		err = json.Unmarshal(body, req)
		return err
	case "application/jce":
		if req, ok := req.(gojce.Message); ok {
			err = req.Decode(bytes.NewBuffer(body))
		}
		return err
	case "application/pb":
		if req, ok := req.(proto.Message); ok {
			err = proto.Unmarshal(body, req)
		}
		return err
	default:
		return fmt.Errorf("invalid Content-Type:%s", head.request.Header["Content-Type"][0])
	}
}
