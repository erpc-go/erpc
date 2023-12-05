package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/erpc-go/erpc/protocol"
	"github.com/erpc-go/erpc/protocol/test"
	"github.com/erpc-go/erpc/utils"
	"github.com/erpc-go/log"
)

// sidecar默认配置
const (
	DefaultEnvoyHost = "127.0.0.1"
	DefaultEnvoyPort = 65001
)

// 应用层协议类型
const (
	AppProtocolTME = "tme" // tme协议
	AppProtocolPDU = "pdu" // pdu协议
	AppProtocolQZA = "qza" // qza协议
)

// context关键属性
var (
	RemoteServiceName = "remote_service_name" // 上游请求中的主调服务名(value为string类型)
	LocalServiceName  = "local_service_name"  // 上游请求中的被调服务名(value为string类型)
	SpanID            = "span_id"             // Span ID(value为uint64类型)
	ParentSpanID      = "parent_span_id"      // Parent Span ID(value为uint64类型)
	TraceID           = "trace_id"            // Trace ID(value为string类型)
	Flag              = "flag"                // Flag(value为uint32类型)
	Env               = "env"                 // env(devops id,string类型)
)

// Client 框架客户端实例
type Client struct {
	Request                    // 网络底层
	reqBody  interface{}       // 请求包
	rspBody  interface{}       // 响应包
	protocol protocol.Protocol // 应用协议首部
	authInfo protocol.AuthInfo // AuthInfo
}

// CallDesc RPC参数
type CallDesc struct {
	LocalServiceName string        // <非必填>本次请求主调服务名
	ServiceName      string        // <必填>本次请求被调服务名, 对应toml配置文件中的一段
	Protocol         string        // <非必填>应用层协议(qza/pdu/tme), 默认tme
	Address          string        // <非必填>
	Timeout          time.Duration // <非必填>RPC超时时间
}

// New 构造支持tme/qdu/qza协议的client
// 底层网络逻辑复用going/client/req
func New(desc CallDesc, authInfo protocol.AuthInfo, reqBody, rspBody interface{}) (c *Client, err error) {
	// step 1. 构造client
	c = &Client{
		reqBody:  reqBody,
		rspBody:  rspBody,
		authInfo: authInfo,
	}

	// step 2. 解析mesh配置
	// serviceName, err := utils.ReplacePattern(desc.ServiceName, "mesh")
	// if err != nil {
	// 	log.Error("serviceName not right")
	// 	return
	// }
	localServiceName, err := utils.ReplacePattern(desc.LocalServiceName, "mesh")
	if err != nil {
		log.Error("localServiceName not right")
		return
	}

	// step 3. 初始化协议首部
	c.protocol = &test.TestProtocol{}
	c.protocol.SetLocalServiceName(localServiceName)

	// step 4. 默认与localhost,65001端口建立tcp长连接
	address := fmt.Sprintf("ip://%s:%d", DefaultEnvoyHost, DefaultEnvoyPort)
	request := Request{
		Network: "tcp",
		ReqType: SendAndRecvKeepalive,
		Address: address,
		Timeout: desc.Timeout,
	}
	if request.Timeout == 0 {
		request.Timeout = 800 * time.Millisecond
	}

	// step 5. 读取动态配置
	// section := desc.ServiceName
	// if serviceName != desc.ServiceName {
	// 	if _, ok := conf[serviceName]; ok {
	// 		section = serviceName
	// 	}
	// }
	// if v, ok := conf[section]; ok {
	// 	// 通信协议
	// 	if len(v.Network) > 0 {
	// 		request.SetNetwork(v.Network)
	// 	}
	// 	// 寻址方式
	// 	if len(v.Address) > 3 {
	// 		request.SetAddress(v.Address)
	// 	}
	// 	// 当前请求延时
	// 	if v.Timeout > 0 {
	// 		request.SetTimeout(v.Timeout)
	// 	}
	// 	// 请求类型
	// 	if 0 != v.ReqType {
	// 		request.SetReqType(v.ReqType)
	// 	}
	// }
	request.Address = desc.Address
	c.Request = request

	return c, nil
}

// Do client单次调用，并且返回错误error
func (c *Client) Do(ctx context.Context, opt ...map[string]string) error {
	defer func() {
		if err := recover(); err != nil {
			log.Error("Recover from panic, err:%+v", err)
		}
	}()

	err := c.doRequest(ctx, opt...)
	if err != nil {
		return err
	}
	// 网络错误则优先返回error
	if c.GetCommuErrCode() != 0 {
		return GetThcErrorFromMsg(int32(c.GetCommuErrCode()), c.GetCommuErrMsg())
	}
	// 然后返回服务器错误
	if c.GetServiceErrCode() != 0 {
		return GetThcErrorFromMsg(int32(c.GetServiceErrCode()), c.GetServiceErrMsg())
	}
	return nil
}

// DoRequest client执行请求
func (c *Client) doRequest(ctx context.Context, opt ...map[string]string) error {
	// Local Service Name
	if localServiceName, ok := ctx.Value(LocalServiceName).(string); ok {
		c.protocol.SetLocalServiceName(localServiceName)
		c.authInfo.CallerInfo = localServiceName
	}
	// Trace ID
	if traceID, ok := ctx.Value(TraceID).(string); ok {
		c.protocol.SetTraceID(traceID)
		c.authInfo.TraceID = traceID
	}
	fmt.Println(c == nil)
	// Remote Serivce Name
	addrs := strings.Split(c.Address, "://")
	if len(addrs) > 1 {
		if addrs[0] == "l5" || addrs[0] == "cl5" || addrs[0] == "gl5" || addrs[0] == "nl5" {
			l5 := strings.Split(addrs[1], ":")
			if len(l5) == 2 {
				c.authInfo.CalleeInfo = "l5-" + l5[0] + "-" + l5[1]
			}
		}
	}
	// Span ID
	if spanID, ok := ctx.Value(SpanID).(uint64); ok {
		c.protocol.SetSpanID(spanID)
	}
	// parent Span ID
	if parentSpanID, ok := ctx.Value(ParentSpanID).(uint64); ok {
		c.protocol.SetParentSpanID(parentSpanID)
	}
	// Flag
	if flag, ok := ctx.Value(Flag).(uint32); ok {
		c.protocol.SetFlag(flag)
	}
	// Env
	if env, ok := ctx.Value(Env).(string); ok {
		c.protocol.SetExtKv(Env, env)
	}
	// 自定义扩展首部
	if len(opt) > 0 {
		for k := range opt[0] {
			c.protocol.SetExtKv(k, opt[0][k])
		}
	}
	fmt.Println(c)
	DoRequests(ctx, c)
	return nil
}

// ReqBody 获取reqbody interface
func (c *Client) ReqBody() interface{} {
	return c.reqBody
}

// RspBody 获取rspbody interface
func (c *Client) RspBody() interface{} {
	return c.rspBody
}

// Marshal 打包函数
func (c *Client) Marshal() ([]byte, error) {
	// if head, ok := c.Head.(*th.TmeHeader); ok {
	// 	// tme协议打包
	// 	bodyBuf, err := th.TMEBodyMarshal(head, c.reqBody)
	// 	if err != nil {
	// 		return bodyBuf, err
	// 	}
	// 	head.SetBodyLen(uint32(len(bodyBuf)))
	// 	pkgBuf := make([]byte, head.Len()+head.GetBodyLen())
	// 	headBuf, _ := head.Marshal()
	// 	copy(pkgBuf[:head.Len()], headBuf)
	// 	copy(pkgBuf[head.Len():], bodyBuf)
	// 	return pkgBuf, nil
	// }
	// if head, ok := c.Head.(*th.QzaHeader); ok {
	// 	// qza协议打包
	// 	bodyBuf, err := th.QZABodyMarshal(head, c.reqBody)
	// 	if err != nil {
	// 		return bodyBuf, err
	// 	}
	// 	head.SetPackLen(uint32(head.GetHeadLen()) + uint32(len(bodyBuf)))
	// 	headBuf, _ := head.Marshal()
	// 	pkgBuf := make([]byte, head.GetPackLen())
	// 	// buffer copy
	// 	copy(pkgBuf[:head.GetPackLen()], headBuf)
	// 	copy(pkgBuf[head.GetHeadLen():], bodyBuf)
	// 	return pkgBuf, nil
	// }
	// if head, ok := c.Head.(*th.PduHeader); ok {
	// 	// pdu协议打包
	// 	bodyBuf, err := th.PDUBodyMarshal(head, c.reqBody)
	// 	if err != nil {
	// 		return bodyBuf, err
	// 	}
	// 	pkgLen := 2 + th.PduProtoHeaderSize + len(bodyBuf) // SOH + 包头 + 包体 + EOT 长度
	// 	head.SetPackLen(uint32(pkgLen))
	// 	headBuf, _ := head.Marshal()
	// 	pkgBuf := make([]byte, pkgLen)
	// 	// buffer copy
	// 	pkgBuf[0] = byte(th.PDUSOH)
	// 	copy(pkgBuf[1:1+th.PduProtoHeaderSize], headBuf)
	// 	copy(pkgBuf[1+th.PduProtoHeaderSize:pkgLen-1], bodyBuf)
	// 	pkgBuf[pkgLen-1] = byte(th.PDUEOT)
	// 	return pkgBuf, nil
	// }
	return nil, fmt.Errorf("Client.Marshal(), invalid protocol")
}

// Unmarshal 解包函数
func (c *Client) Unmarshal(data []byte) error {
	// if head, ok := c.Head.(*th.TmeHeader); ok {
	// 	// tme协议解包
	// 	err := head.Unmarshal(data)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	c.ServiceErrCode = int(head.GetResultCode())
	// 	if c.rspBody != nil && head.GetBodyLen() > 0 {
	// 		return th.TMEBodyUnmarshal(head, data, c.rspBody)
	// 	}
	// 	return nil
	// }
	// if head, ok := c.Head.(*th.QzaHeader); ok {
	// 	// qza协议解包
	// 	err := head.Unmarshal(data)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	c.ServiceErrCode = int(head.GetResultCode())
	// 	if c.rspBody != nil && head.GetBodyStartIndex() < int16(len(data)) {
	// 		return th.QZABodyUnmarshal(head, data, c.rspBody)
	// 	}
	// 	return nil
	// }
	// if head, ok := c.Head.(*th.PduHeader); ok {
	// 	// pdu协议解包
	// 	err := head.Unmarshal(data)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	c.ServiceErrCode = int(head.GetResultCode())
	// 	if c.rspBody != nil && len(data) > 25 {
	// 		return th.PDUBodyUnmarshal(head, data, c.rspBody)
	// 	}
	// 	return nil
	// }
	return nil
}

// GetLastCallee 获取被调服务信息
func (c *Client) GetLastCallee() string {
	return ""
}

// Check 包完整性校验
func (c *Client) Check(data []byte) (int, error) {
	return len(data), nil
}

func isValidService(service string) (b bool) {
	return
}
