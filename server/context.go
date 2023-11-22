package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"runtime"
	"time"

	"github.com/erpc-go/erpc/protocol"
	"github.com/erpc-go/jce-codec"
	"github.com/erpc-go/log"
)

var (
	RemoteServiceName = "remote_service_name" // 主调服务名(value为string类型)
	LocalServiceName  = "local_service_name"  // 被调服务名(value为string类型)
	SpanID            = "span_id"             // Span ID(value为uint64类型)
	ParentSpanID      = "parent_span_id"      // Parent Span ID(value为uint64类型)
	TraceID           = "trace_id"            // Trace ID(value为string类型)
	Flag              = "flag"                // Flag(value为uint32类型)
	Env               = "env"                 // Env(Devops环境,string类型)
)

// Context 上下文
type Context struct {
	Req        jce.Messager // 请求包
	Rsp        jce.Messager // 响应包
	startTime  time.Time    // 创建时间
	endTime    time.Time    //
	noResponse bool         // 是否需要回包
	Protocol   protocol.Protocol
	jsonFormat bool // response是否进行格式化输出
	fromWNS    bool // 是否WNS协议

	context.Context
	Uin           uint64
	Cmd           uint32
	SubCmd        uint32
	Seq           uint32
	ClientAddr    uint32
	ClientVersion uint32
	ClientIP      net.IP
	ErrCode       int32
	ExtData       interface{} // 额外数据，用于异步逻辑传参
	LogLevel      int         // 日志打印等级，每次new一个context时，都必须设置这个
	Version       int         // 客户端版本号
	Source        int         // 客户端来源 1:ios 2:android 3:web
	writer        io.Writer
	buffer        *bytes.Buffer
	level         int
}

// NewContext 创建新上下文
func NewContext(ctx context.Context) *Context {
	newCtx := Context{
		startTime: time.Now(),
	}
	newCtx.LogLevel = 0
	return &newCtx
}

// Now 返回请求进入时间 return enter time, no need to call time.Now() every time , every where
func (ctx *Context) Now() time.Time {
	return ctx.startTime
}

// Cost 返回当前耗时 return cost time
func (ctx *Context) Cost() time.Duration {
	ctx.endTime = time.Now()
	return time.Now().Sub(ctx.startTime)
}

// Result 获取此次请求返回码
func (ctx *Context) Result() int32 {
	if ctx.Protocol == nil {
		return 0
	}
	return ctx.Protocol.GetResultCode()
}

// GetResult getter应该去掉Get前缀,
// 这里由于已经有服务使用GetResult(),所以先保留
func (ctx *Context) GetResult() int32 {
	return ctx.Result()
}

// SetResult 设置此次请求返回码
func (ctx *Context) SetResult(resultCode int32) {
	if ctx.Protocol == nil {
		return
	}
	ctx.Protocol.SetResultCode(resultCode)
}

// ResultMsg 获取错误信息
func (ctx *Context) ResultMsg() string {
	if ctx.Protocol == nil {
		return ""
	}
	return ctx.Protocol.GetResultMsg()
}

// SetResultMsg 设置错误信息
func (ctx *Context) SetResultMsg(msg string) {
	if ctx.Protocol == nil {
		return
	}
	ctx.Protocol.SetResultMsg(msg)
}

// NoResponse 获取此次请求是否需要回包
func (ctx *Context) NoResponse() bool {
	return ctx.noResponse
}

// SetNoResponse 设置该次请求不需要回包
func (ctx *Context) SetNoResponse() {
	ctx.noResponse = true
}

// RemoteServiceName 获取主调服务名(上游请求th中的LocalServiceName)
func (ctx *Context) RemoteServiceName() string {
	if remoteServiceName, ok := ctx.Value(RemoteServiceName).(string); ok {
		return remoteServiceName
	}
	return ""
}

// LocalServiceName 获取被调服务名(上游请求th中的ServiceName)
func (ctx *Context) LocalServiceName() string {
	if localServiceName, ok := ctx.Value(LocalServiceName).(string); ok {
		return localServiceName
	}
	return ""
}

// 【已废弃, 不建议继续使用】获取被调服务名(上游请求th中的ServiceName)
func (ctx *Context) ServiceName() string {
	return ctx.LocalServiceName()
}

// RemoteAddr 获取主调服务ip:port
func (ctx *Context) RemoteAddr() string {
	// if remoteAddr, ok := ctx.Value(cat.ClientAddr).(string); ok {
	// 	return remoteAddr
	// }
	return ""
}

// LocalAddr 获取被调服务ip:port
func (ctx *Context) LocalAddr() string {
	// if localAddr, ok := ctx.Value(cat.ServerAddr).(string); ok {
	// 	return localAddr
	// }
	return ""
}

// GetAuthInfo 获取登录态
func (ctx *Context) GetAuthInfo() protocol.AuthInfo {
	if ctx.Protocol == nil {
		return protocol.AuthInfo{}
	}
	return ctx.Protocol.GetAuthInfo()
}

// TraceID 获取链路ID
func (ctx *Context) TraceID() string {
	if traceID, ok := ctx.Value(TraceID).(string); ok {
		return traceID
	}
	return ""
}

// SpanID 获取节点ID
func (ctx *Context) SpanID() uint64 {
	if spanID, ok := ctx.Value(SpanID).(uint64); ok {
		return spanID
	}
	return 0
}

// ParentSpanID 获取父节点ID
func (ctx *Context) ParentSpanID() uint64 {
	if parentSpanID, ok := ctx.Value(ParentSpanID).(uint64); ok {
		return parentSpanID
	}
	return 0
}

// Flag 获取染色标志
func (ctx *Context) Flag() uint32 {
	if flag, ok := ctx.Value(Flag).(uint32); ok {
		return flag
	}
	return 0
}

// Env 获取DevOps环境ID
func (ctx *Context) Env() string {
	if env, ok := ctx.Value(Env).(string); ok {
		return env
	}
	return ""
}

// UID 获取用户ID
func (ctx *Context) UID() uint64 {
	if ctx.Protocol == nil {
		return 0
	}
	return ctx.Protocol.GetUid()
}

// AppID 获取APP编号
func (ctx *Context) AppID() uint32 {
	if ctx.Protocol == nil {
		return 0
	}
	return ctx.Protocol.GetAppID()
}

func (ctx *Context) FromWNS() bool {
	return ctx.fromWNS
}

func (ctx *Context) SetFromWNS(fromWNS bool) {
	ctx.fromWNS = fromWNS
}

// JSONFormat 获取是否格式化JSON标记
func (ctx *Context) JSONFormat() bool {
	return ctx.jsonFormat
}

// SetJSONFormat 设置是否格式话JSON标记
func (ctx *Context) SetJSONFormat(b bool) {
	ctx.jsonFormat = b
}

// 获取Extend扩展首部
func (ctx *Context) Extends() map[string]string {
	if ctx.Protocol == nil {
		return nil
	}
	return ctx.Protocol.GetExtends()
}

// 获取统一日志需要的字段
func (ctx *Context) LogTraceMetadata() map[string]string {
	return map[string]string{
		"ServiceName":  ctx.LocalServiceName(),
		"TraceID":      ctx.TraceID(),
		"SpanID":       fmt.Sprintf("%d", ctx.SpanID()),
		"ParentSpanID": fmt.Sprintf("%d", ctx.ParentSpanID()),
		"Flag":         fmt.Sprintf("%d", ctx.Flag()),
		"UID":          fmt.Sprintf("%d", ctx.UID()),
		"AppID":        fmt.Sprintf("%d", ctx.AppID()),
	}
}

// AsyncProcess 回包后异步作业
func (ctx *Context) AsyncProcess(f func(*Context)) {
	baseCtx := context.Background()
	// context赋值(remote_service_name, local_service_name, trace_id, span_id, parent_span_id, env)
	baseCtx = context.WithValue(baseCtx, RemoteServiceName, ctx.RemoteServiceName())
	baseCtx = context.WithValue(baseCtx, LocalServiceName, ctx.LocalServiceName())
	baseCtx = context.WithValue(baseCtx, TraceID, ctx.TraceID())
	baseCtx = context.WithValue(baseCtx, SpanID, ctx.SpanID())
	baseCtx = context.WithValue(baseCtx, ParentSpanID, ctx.ParentSpanID())
	baseCtx = context.WithValue(baseCtx, Flag, ctx.Flag())
	baseCtx = context.WithValue(baseCtx, Env, ctx.Env())
	c := NewContext(baseCtx)
	if ctx.Protocol != nil {
		c.Protocol = ctx.Protocol.Clone()
	}
	c.ExtData = ctx.ExtData
	// c.Req = JceMsgClone(ctx.Req)
	// c.Rsp = JceMsgClone(ctx.Rsp)

	go func() {
		// defer c.WriteLog()
		defer func() {
			if e := recover(); e != nil {
				buf := make([]byte, 4*1024*1024)
				buf = buf[:runtime.Stack(buf, false)]
				log.Raw(">> %s", e, buf)
			}
		}()
		f(c)
	}()
}
