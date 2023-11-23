package client

import (
	"context"
	"fmt"
	"log"
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	uint64seq uint64 = 1
	uint32Seq uint32 = 1
	bufPool          = sync.Pool{New: func() interface{} { return make([]byte, maxRspDataLen) }}
)

// NewUint32Seq 生成全局唯一的uint32 seq
func NewUint32Seq() uint32 {
	return atomic.AddUint32(&uint32Seq, 1)
}

// NewUint64Seq 生成全局唯一的uint64 seq
func NewUint64Seq() uint64 {
	return atomic.AddUint64(&uint64seq, 1)
}

// 后端网络请求错误码
const (
	ErrOK              = 0  // 成功
	ErrRequesting      = 1  // 在还未发起请求前进行检查，检查不通过，例如缺少必须参数、缺少请求或返回体等
	ErrReqInfoInvalid  = 2  // DSN配置错误
	ErrAddressInvalid  = 3  // address配置错误
	ErrAddressingFail  = 4  // 没有安装cmlb l5客户端
	ErrNetworkInvalid  = 5  // 只支持 tcp udp unix zmq
	ErrResolveAddrFail = 6  // address配置ip格式有问题
	ErrDialConnFail    = 7  // 一般网络不通, 或者后端core，没有进程, 或者端口不够用
	ErrMarshalFail     = 8  // 打包失败
	ErrSendFail        = 9  // 发包失败
	ErrSendTimeout     = 10 // 发包超时
	ErrRecvFail        = 11 // 收包失败 一般是tcp对端主动断开连接
	ErrRspDataTooLarge = 12 // udp包最大64k， tcp包最大64M
	ErrRecvTimeout     = 13 // 收包超时，这个是最多的失败情况
	ErrCheckFail       = 14 // 验包失败, 一般是非法包串入
	ErrUnmarshalFail   = 15 // 解包失败
	ErrRequestPanic    = 16 // 请求后端panic， 一般是打解包空指针问题
	ErrContextCanceled = 17 // http/tcp主动断开连接，会cancel掉context，提前结束，避免浪费资源
	ErrContextTimeout  = 18 // 全局超时
)

// 后端请求类型 request type
const (
	SendAndRecv            = 1 // 普通的一来一回req/rsp
	SendAndRecvKeepalive   = 2 // 使用tcp连接池
	SendOnlyKeepalive      = 3 // tcp连接池只发不收
	SendOnly               = 4 // 只发不收
	SendAndRecvIgnoreError = 5 // web长轮询技术，没有消息返回超时是正常的，应该忽略不上报l5
	SendAndRecvStream      = 6 // tcp stream transport
	SendOnlyStream         = 7 // tcp流一直发
)

// ReqTypeMsg 后端网络请求类型字面量
var ReqTypeMsg = map[int]string{
	0: "Unknown",
	1: "SendAndRecv",
	2: "SendAndRecvKeepalive",
	3: "SendOnlyKeepalive",
	4: "SendOnly",
	5: "SendAndRecvIgnoreError",
	6: "SendAndRecvStream",
	7: "SendOnlyStream",
}

const (
	maxRspDataLen              = 65536 // 64k
	retryTimesWhenUDPCheckFail = 1     // udp验包失败的重试次数，防止串包,野包
)

// Requestor 后端请求需要实现的接口 an interface that client uses to marshal/unmarshal, and then request
type Requestor interface {
	DataSourceName() string // DataSourceName cmlb://appid?timeout=300&reqtype=1&network=udp
	Cmd() string
	Marshal() ([]byte, error)
	Check([]byte) (int, error)
	Unmarshal([]byte) error
	Finish(errcode int, address string, cost time.Duration) // Finish return error code, address, cost time when request finish
}

// DoRequests 多并发请求
func DoRequests(ctx context.Context, reqs ...Requestor) {
	done := isDone(ctx)

	if len(reqs) == 1 {
		req := reqs[0]
		if done > 0 {
			finish(ctx, req, done, "", 0)
			return
		}

		reqInfo := NewReqInfoFromDSN(req.DataSourceName())
		c, f := context.WithTimeout(ctx, reqInfo.Timeout)
		doRequest(c, req, reqInfo)
		f()
	} else {
		var wg sync.WaitGroup
		for _, req := range reqs {
			if done > 0 {
				finish(ctx, req, done, "", 0)
				return
			}

			wg.Add(1)
			reqInfo := NewReqInfoFromDSN(req.DataSourceName())
			subCtx, cancel := context.WithTimeout(ctx, reqInfo.Timeout)
			go func(r Requestor, c context.Context, f context.CancelFunc, info *ReqInfo) {
				doRequest(c, r, info)
				f()
				wg.Done()
			}(req, subCtx, cancel, reqInfo)
		}
		wg.Wait()
	}
}

func finish(ctx context.Context, req Requestor, errcode int, address string, cost time.Duration) {
	req.Finish(errcode, address, cost)
}

func isDone(ctx context.Context) int {
	select {
	case <-ctx.Done():
		if ctx.Err() == context.Canceled {
			return ErrContextCanceled
		}
		if ctx.Err() == context.DeadlineExceeded {
			return ErrContextTimeout
		}
		return 0
	default:
	}

	return 0
}

func doRequest(ctx context.Context, r Requestor, reqInfo *ReqInfo) {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 16*1024*1024)
			buf = buf[:runtime.Stack(buf, false)]
			log.Output(1, fmt.Sprintf("[PANIC]%v\n%s", err, buf))
			finish(ctx, r, ErrRequestPanic, "", 0)
		}
	}()

	addressing, err := NewAddress(reqInfo.Address)
	if err != nil {
		log.Output(1, fmt.Sprintf("addressing:%s fail:%s", reqInfo.Address, err))
		finish(ctx, r, ErrAddressingFail, "", 0)
		return
	}
	addr := addressing.Address()

	// check if done after addressing
	if done := isDone(ctx); done > 0 {
		finish(ctx, r, done, addr, addressing.Cost())
		return
	}

	var ec int
	if reqInfo.Network == "udp" {
		ec = doUDPRequest(ctx, r, addr, reqInfo)
	} else if reqInfo.Network == "tcp" {
		ec = doNetworkRequest(ctx, r, addr, reqInfo)
		if reqInfo.ReqType == SendAndRecvKeepalive && ec == ErrRecvFail {
			ec = doNetworkRequest(ctx, r, addr, reqInfo)
		}
	} else if reqInfo.Network == "unix" {
		ec = doNetworkRequest(ctx, r, addr, reqInfo)
	} else {
		finish(ctx, r, ErrNetworkInvalid, addr, addressing.Cost())
		return
	}
	finish(ctx, r, ec, addr, addressing.Cost())

	if reqInfo.ReqType == SendAndRecvIgnoreError {
		ec = ErrOK
	}
	if ec == ErrOK {
	} else {
	}

	err = nil
	if ec == ErrDialConnFail || ec == ErrRecvTimeout || ec == ErrRecvFail {
		err = fmt.Errorf("[%d,%s]", ec, ErrMsg[ec])
	}
	addressing.Update(err)
}

// doNetworkRequest tcp 网络请求
func doNetworkRequest(ctx context.Context, r Requestor, addr string, reqInfo *ReqInfo) int {
	d, _ := ctx.Deadline()
	timeout := d.Sub(time.Now())

	var conn net.Conn
	var err error
	var pool *Pool
	shouldReturnPool := false
	if reqInfo.ReqType == SendAndRecvKeepalive || reqInfo.ReqType == SendOnlyKeepalive {
		key := fmt.Sprintf("%s:%d", addr, reqInfo.ReqType) // ip:port:1
		pool = GetTCPConnectionPool(key, addr, reqInfo.Network, timeout)

		c, e := pool.Get()
		if c == nil {
			return ErrDialConnFail
		}
		if e != nil {
			return ErrDialConnFail
		}

		var ok bool
		if conn, ok = c.(net.Conn); !ok {
			return ErrDialConnFail
		}
	} else {
		conn, err = net.DialTimeout(reqInfo.Network, addr, timeout)
		if err != nil {
			return ErrDialConnFail
		}
	}
	defer func() {
		if shouldReturnPool && pool != nil {
			pool.Put(conn)
		} else {
			conn.Close()
			if pool != nil && !shouldReturnPool {
			}
		}
	}()

	// check if done after dial
	if done := isDone(ctx); done > 0 {
		return done
	}

	d, _ = ctx.Deadline()
	conn.SetDeadline(d)

	reqData, err := r.Marshal()
	if err != nil || len(reqData) == 0 {
		log.Output(1, fmt.Sprintf("marshal fail:%s, req data len:%d", err, len(reqData)))
		return ErrMarshalFail
	}

	sentNum := 0
	for sentNum < len(reqData) {
		var num int
		num, err = conn.Write(reqData[sentNum:])
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				return ErrSendTimeout
			}
			log.Output(1, fmt.Sprintf("send fail:%v", err))
			return ErrSendFail
		}
		sentNum += num

		// check if done after write
		if done := isDone(ctx); done > 0 {
			return done
		}
	}

	if reqInfo.ReqType == SendOnlyKeepalive || reqInfo.ReqType == SendOnly {
		shouldReturnPool = true
		return ErrOK
	}

	buf := bufPool.Get()
	rspData, _ := buf.([]byte)
	defer func() {
		bufPool.Put(rspData) // tcp包过大扩充时会重新赋值，所以defer必须放在闭包里面
	}()

	recvNum := 0
	checkNum := 0
	for {
		var num int
		num, err = conn.Read(rspData[recvNum:])
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				return ErrRecvTimeout
			}
			log.Output(1, fmt.Sprintf("recv fail:%v", err))
			return ErrRecvFail
		}
		recvNum += num
		if recvNum >= cap(rspData) {
			if recvNum >= 1024*maxRspDataLen {
				fmt.Println("recv rsp data too big, larger than 64M, return fail, recv num:", recvNum)
				return ErrRspDataTooLarge
			}
			fmt.Println("recv rsp data too big, expand twice cap, recv num:", recvNum)
			tmpRspData := make([]byte, recvNum*2)
			copy(tmpRspData, rspData[:recvNum])
			rspData = tmpRspData
		}

		// check if done after read
		if done := isDone(ctx); done > 0 {
			return done
		}

		checkNum, err = r.Check(rspData[:recvNum])
		if err != nil || checkNum < 0 {
			return ErrCheckFail
		}
		if checkNum > 0 {
			if checkNum < recvNum {
			}
			if checkNum > recvNum {
				return ErrCheckFail
			}
			break
		}
	}

	err = r.Unmarshal(rspData[:checkNum])
	if err != nil {
		log.Output(1, fmt.Sprintf("unmarshal fail:%s", err))
		return ErrUnmarshalFail
	}

	shouldReturnPool = true
	return ErrOK
}

// doUDPRequest udp客户端请求类似udp服务端 解决回包路径不一致问题 比如:  req: A -> B -> C, rsp: C -> A (imagent)
func doUDPRequest(ctx context.Context, r Requestor, addr string, reqInfo *ReqInfo) int {
	conn, err := net.ListenPacket("udp4", ":") // 直接listen udp， 对于多网卡会有bug
	if err != nil {
		log.Output(1, "listen udp packet fail:"+err.Error())
		return ErrDialConnFail
	}
	defer conn.Close()

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Output(1, "resolve udp addr fail:"+err.Error())
		return ErrResolveAddrFail
	}

	// check if done after listen packet
	if done := isDone(ctx); done > 0 {
		return done
	}

	d, _ := ctx.Deadline()
	conn.SetDeadline(d)

	reqData, err := r.Marshal()
	if err != nil || len(reqData) == 0 {
		log.Output(1, fmt.Sprintf("marshal fail:%s, req data len:%d", err, len(reqData)))
		return ErrMarshalFail
	}

	sentNum, err := conn.WriteTo(reqData, udpAddr)
	if err != nil || sentNum == 0 {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return ErrSendTimeout
		}
		log.Output(1, fmt.Sprintf("send fail:%v", err))
		return ErrSendFail
	}

	// check if done after WriteTo
	if done := isDone(ctx); done > 0 {
		return done
	}

	if reqInfo.ReqType == SendOnly {
		return ErrOK
	}

	buf := bufPool.Get()
	rspData, _ := buf.([]byte)
	defer bufPool.Put(rspData)

	tryTimes := retryTimesWhenUDPCheckFail + 1
	recvNum := 0
	checkNum := 0
	for tryTimes > 0 {
		tryTimes--

		recvNum, _, err = conn.ReadFrom(rspData)
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				return ErrRecvTimeout
			}
			log.Output(1, fmt.Sprintf("recv fail:%v", err))
			return ErrRecvFail
		}

		// check if done after read
		if done := isDone(ctx); done > 0 {
			return done
		}

		checkNum, err = r.Check(rspData[:recvNum])
		if err != nil || checkNum <= 0 {
			if tryTimes > 0 {
				continue
			} else {
				return ErrCheckFail
			}
		} else {
			break
		}
	}

	err = r.Unmarshal(rspData[:recvNum])
	if err != nil {
		log.Output(1, "unmarshal fail:"+err.Error())
		return ErrUnmarshalFail
	}

	return ErrOK
}

// ErrMsg 后端网络请求错误信息
var ErrMsg = map[int]string{
	0:  "OK",
	1:  "Requesting",
	2:  "ReqinfoInvalid",
	3:  "AddressInvalid",
	4:  "AddressingFail",
	5:  "NetworkInvalid",
	6:  "ResolveFail",
	7:  "DialFail",
	8:  "MarshalFail",
	9:  "SendFail",
	10: "SendTimeout",
	11: "RecvFail",
	12: "RspPkgTooBig",
	13: "RecvTimeout",
	14: "CheckFail",
	15: "UnmarshalFail",
	16: "RequestPanic",
	17: "ContextCanceled",
	18: "ContextTimeout",
	19: "Unknown",
}

// ReqInfo 后端请求必需信息 由DataSourceName解析出来
type ReqInfo struct {
	Network string        // tcp udp unix zmq
	Address string        // l5://modid:cmdid  ip://ip:port  cmlb://appid dns://id.qq.com:80  sock://filepath
	ReqType int           // request type: SendAndRecv SendAndRecvKeepalive SendOnlyKeepalive SendOnly SendAndRecvIgnoreError
	Timeout time.Duration // current action timeout time.Second
	ZmqNet  string        // zmq only: tcp inproc
}

var (
	reqInfoMap  = make(map[string]*ReqInfo, 0)
	reqInfoLock sync.RWMutex
)

// NewReqInfoFromDSN 由DSN生成ReqInfo get req info from data source name: cmlb://appid?timeout=300&reqtype=1&network=udp
func NewReqInfoFromDSN(dsn string) *ReqInfo {
	if len(dsn) > 3 && dsn[:3] == "cl5" { // cl5一致性哈希是动态寻址，不可缓存
		r := &ReqInfo{
			Network: "udp",
			ReqType: SendAndRecv,
			Timeout: 800 * time.Millisecond,
			ZmqNet:  "tcp",
		}
		addrs := strings.Split(dsn, "?")
		if len(addrs) != 2 {
			return r
		}
		r.Address = addrs[0]

		if len(addrs) > 1 {
			params := strings.Split(addrs[1], "&")
			for _, val := range params {
				p := strings.Split(val, "=")
				if len(p) != 2 {
					continue
				}
				if p[0] == "timeout" {
					timeout, _ := strconv.Atoi(p[1])
					r.Timeout = time.Duration(timeout) * time.Millisecond
					continue
				}
				if p[0] == "reqtype" {
					reqtype, _ := strconv.Atoi(p[1])
					r.ReqType = reqtype
					continue
				}
				if p[0] == "network" {
					r.Network = p[1]
					continue
				}
				if p[0] == "zmqnet" {
					r.ZmqNet = p[1]
					continue
				}
			}
		}
		return r
	}

	reqInfoLock.RLock()
	r, ok := reqInfoMap[dsn]
	reqInfoLock.RUnlock()
	if ok {
		return r
	}
	reqInfoLock.Lock()
	defer reqInfoLock.Unlock()
	r, ok = reqInfoMap[dsn]
	if ok {
		return r
	}
	r = &ReqInfo{
		Network: "udp",
		Address: "",
		ReqType: SendAndRecv,
		Timeout: 800 * time.Millisecond,
		ZmqNet:  "tcp",
	}
	reqInfoMap[dsn] = r

	addrs := strings.Split(dsn, "?")
	if len(addrs) > 0 {
		r.Address = addrs[0]
	}

	if len(addrs) > 1 {
		params := strings.Split(addrs[1], "&")
		for _, val := range params {
			p := strings.Split(val, "=")
			if len(p) != 2 {
				continue
			}
			if p[0] == "timeout" {
				timeout, _ := strconv.Atoi(p[1])
				r.Timeout = time.Duration(timeout) * time.Millisecond
				continue
			}
			if p[0] == "reqtype" {
				reqtype, _ := strconv.Atoi(p[1])
				r.ReqType = reqtype
				continue
			}
			if p[0] == "network" {
				r.Network = p[1]
				continue
			}
			if p[0] == "zmqnet" {
				r.ZmqNet = p[1]
				continue
			}
		}
	}
	return r
}
