package server

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/erpc-go/erpc/protocol"
	"github.com/erpc-go/erpc/server/net"
	"github.com/erpc-go/jce-codec"
	"github.com/erpc-go/log"
	limit "github.com/erpc-go/ratelimit"
)

const (
	stackSize = 4 * 1024 * 1024
)

type mutexEntry struct {
	h       Handler
	pattern string
	token   string
	reqType jce.Messager
	rspType jce.Messager
}

// ServerMutex 带读写锁的server入口配置
type ServeMutex struct {
	mutex      sync.RWMutex
	mapEntries map[string]mutexEntry

	Addr                  string        // 网卡:端口/协议，0.0.0.0对应的网卡是all, 如 eth1:10100/udp
	Name                  string        `default:"going-svr"` // 服务名字
	User                  string        `default:"going"`     // 服务负责人
	MsgTimeout            time.Duration `default:"800ms"`     // 当前请求全局超时时间，默认800ms
	IdleTimeout           time.Duration `default:"3m"`        // tcp server长链接最大空闲时间，默认3min
	EnableGracefulRestart bool          `default:"true"`      // 是否支持热重启
	MaxWorkerCount        int           `default:"10000"`     // 协程池最大协程数，并发请求数，用于过载保护
	Ratelimit             int64         // 限频，默认不开启
	EnableDebugMode       bool          // 开启调试模式，打印更详细日志
	MasterID              int           // 模调主调模块id
	StartAttr             int           // 启动量监控属性id,通过 going attr config.toml自动生成
	PanicAttr             int           // panic次数
	EnterAttr             int           // 进入量
	SuccAttr              int           // 成功量
	FailAttr              int           // 失败量
	LogicFailAttr         [][]int       // 逻辑失败属性，这里配置了就不会上报到FailAttr, [[errcode, attrid], [1, 2], [3, 4]]
	CostAttr200           int           // 耗时小于200ms的请求量
	CostAttr800           int           // 耗时200-800ms的请求量
	CostAttr800p          int           // 耗时大于800ms的请求量
	GoroutineCountAttr    int           // goroutine协程数 属性必须是时刻量
	ThreadCountAttr       int           // thread线程数 属性必须是时刻量
	AllocHeapAttr         int           // Alloc已分配且在使用中的内存字节数(单位:M) 属性必须是时刻量
	NumGCAttr             int           // NumGC已经完成的GC循环次数(单位:千) 属性必须是时刻量
	PauseTotalNsAttr      int           // PauseTotalNs自程序启动后的GC总暂停时间(单位:秒) 属性必须是时刻量
	PauseNsAttr           int           // PauseNs最近256次GC的平均暂停时间(单位:纳秒) 属性必须是时刻量
	LogicFailAttrMap      map[int]int   // 这里不允许配置，通过LogicFailAttr生成

	ListenIP   string // 通过解析addr生成
	ListenPort uint16 // 通过解析addr生成
	ListenNet  string // tcp udp all
	Address    string // 通过解析addr生成 ip:port
}

// HandleFunc 自动转化成Handler即可
func (sm *ServeMutex) HandleFunc(pattern, token string, handler HandlerFunc, reqType, rspType jce.Messager) {
	sm.Handle(pattern, token, handler, reqType, rspType)
}

// Handle 注册相应的cmd pattern，token和处理函数到map中
func (sm *ServeMutex) Handle(pattern, token string, handler Handler, reqType, rspType jce.Messager) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if handler == nil {
		panic("invalid handler")
	}

	if sm.mapEntries == nil {
		sm.mapEntries = make(map[string]mutexEntry)
	}

	// 命令字通配
	pattern, err := ReplacePattern(pattern, "mesh")
	if err != nil {
		panic("invalid pattern")
	}

	sm.mapEntries[pattern] = mutexEntry{
		h:       handler,
		pattern: pattern,
		token:   token,
		reqType: reqType,
		rspType: rspType,
	}
}

// Alias 服务别名绑定，将src服务名对应的func依次绑定到dst对应的服务名上
func (sm *ServeMutex) Alias(src string, dst ...string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 命令字通配
	src, err := ReplacePattern("")
	if err != nil {
		panic("invalid pattern")
	}

	if _, ok := sm.mapEntries[src]; !ok {
		panic("pattern not find")
	}

	for _, v := range dst {
		entry := mutexEntry{
			h:       sm.mapEntries[src].h,
			pattern: v,
			token:   "",
			reqType: sm.mapEntries[src].reqType,
			rspType: sm.mapEntries[src].rspType,
		}
		sm.mapEntries[v] = entry
	}
}

func (sm *ServeMutex) Serve(baseCtx context.Context, reqBuf []byte) (b []byte, err error) {
	// step1
	defer func() {
		if e := recover(); e != nil {
			dataBuf := make([]byte, stackSize)
			dataBuf = dataBuf[:runtime.Stack(dataBuf, false)]
			log.Panic("%v\n>> %s", e, dataBuf)
		}
	}()

	pt := protocol.GetProtocolType(reqBuf)
	p := GetProtocolStruct(pt)
	if err := p.UnmarshalHeader(reqBuf); err != nil {
		log.Raw("%s header Unmarshal failed, msg:%v", pt.String(), err)
		return nil, err
	}

	ctx := NewContext(baseCtx)

	// 设置协议首部
	ctx.Protocol = p

	// 映射handler
	entry, ok := sm.mapEntries[p.GetCmdPattern()]
	if !ok {
		log.Raw("cmd pattern[%s] not find the entry!", p.GetCmdPattern())
		return nil, fmt.Errorf("invalid cmd patrern:%s", p.GetCmdPattern())
	}

	// 应用协议解析
	ctx.Req = entry.reqType
	ctx.Rsp = entry.rspType

	// body
	if err := p.UnmarshalBody(reqBuf, ctx.Req); err != nil {
		log.Raw("protocol body Decode buf failed, msg:%v", err)
		return nil, err
	}

	// 业务处理逻辑
	entry.h.Process(ctx)

	// 处理耗时
	ctx.Cost()

	// 返回码
	ctx.ErrCode = ctx.Protocol.GetResultCode()

	// 不回包
	if ctx.NoResponse() {
		return nil, errors.New("no response")
	}

	// 打包
	bodyBuf, err := p.MarshalBody(ctx.Rsp)
	if err != nil {
		log.Raw("protocol body Decode buf failed, msg:%v", err)
		return nil, err
	}
	p.SetBodyLen(uint32(len(bodyBuf)))
	headBuf, err := p.MarshalHeader()
	if err != nil {
		log.Raw("protocol header marshal buf failed, msg:%v", err)
		return nil, err
	}
	pkgBuf := make([]byte, len(headBuf)+len(bodyBuf))
	copy(pkgBuf[:len(headBuf)], headBuf)
	copy(pkgBuf[len(headBuf):], bodyBuf)
	return pkgBuf, nil
}

func (sm *ServeMutex) Listen() {
	// 初始化
	log.Raw("==>-----------------erpc start at %s-----------------\n==>\n", time.Now())

	// 速率限制，按请求个数
	var limiter limit.Limiter
	limiter = limit.New(999)

	// 解析环境变量,获取父进程已注册服务列表
	parentServices := ParseServiceFromEnv()
	log.Raw("parent' services:%+v\n", parentServices)
	// 当前进程待写入环境变量的服务列表
	var registeredServices []*Service
	// 进程停止信号
	stop := make(chan struct{})

	// 遍历父进程已注册服务列表
	// 若子进程无此服务/前进程更换了端口, 则进行解注册
	// 若子进程存在此服务, 不用处理
	// 若子进程有新服务, 进行注册
	mapEntries := make(map[string]mutexEntry) // 此处一定要deep copy
	for k, v := range sm.mapEntries {
		if v.token == "" {
			continue
		}
		mapEntries[k] = v
	}

	if 0 != len(parentServices) {
		chanRegisterdServices := make(chan *Service, len(parentServices))
		for _, value := range parentServices {
			// 子进程更新了端口
			if int(sm.ListenPort) != value.Port {
				log.Raw("[service register]Current Process's Port Has Changed. But Restart Graceful Keep The Old One.%+v\n", value)
			}
			// 子进程无此服务
			if _, ok := mapEntries[value.Name]; !ok {
				chanRegisterdServices <- value
				log.Raw("[service register]Current Process Hasn't This Service, %+v\n", value)
			} else {
				// 子进程存在此服务,不需要注册
				log.Raw("[service register]Current Process Has This Service, %+v\n", value)
				registeredServices = append(registeredServices, value)
				delete(mapEntries, value.Name)
			}
		}
		// serviceDeregister(chanRegisterdServices)
	}
	log.Raw("[service register]Need Register Service Size:%d\n", len(mapEntries))

	// 子进程需进行服务注册的channel
	chanNeedRegister := make(chan *Service, len(mapEntries)+len(registeredServices))

	// 服务注册异步调用
	go func() {
		// 延迟启动时间支持可配,默认5s
		if 0 == sm.MsgTimeout {
			time.Sleep(5 * time.Second)
		} else {
			time.Sleep(sm.IdleTimeout)
		}
		// serviceRegister(stop, registeredServices, mapEntries, chanNeedRegister)
	}()

	// 监听信号
	go handleSignal(chanNeedRegister, stop)

	log.Raw("[handlers]%+v\n", sm.mapEntries)

	// 端口监听
	net.Init(sm.MsgTimeout, sm.IdleTimeout, sm.EnableDebugMode, 0, 0, false)
	switch sm.ListenNet {
	case "tcp":
		net.ListenAndServeTCP(sm.Address, net.CheckerFunc(Check), sm, limiter)
	case "udp":
		net.ListenAndServeUDP(sm.Address, net.CheckerFunc(Check), sm, limiter)
	case "all":
		net.ListenAndServe(sm.Address, net.CheckerFunc(Check), sm, limiter)
	default:
		panic("invalid listening network config!")
	}
}

func isValidPattern(p string) (b bool) {
	return true
}

func ReplacePattern(in string, replaces ...string) (string, error) {
	if strings.Index(in, "/") != -1 {
		return in, nil
	}
	s := strings.Split(in, ".")
	if len(s) != 4 {
		return "", fmt.Errorf("%s need 4 sections", in)
	}
	for i := range s {
		if s[i] != "*" {
			continue
		}
		if len(replaces) <= i || len(replaces[i]) == 0 {
			return "", fmt.Errorf("%s's %d section invalid", in, i)
		}
		s[i] = replaces[i]
	}
	out := s[0]
	for i := 1; i < len(s); i++ {
		out += "." + s[i]
	}

	return out, nil
}
