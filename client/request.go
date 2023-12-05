package client

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Conf common client conf
type Conf struct {
	Address  string
	Network  string        `default:"udp"`
	ReqType  int           `default:"1"`
	Timeout  time.Duration `default:"800ms"`
	Sidecar  bool          // 是否开启sidecar模式
	Command  string        // 后端请求命令字
	Password string        // redis password
	Test     int           // 配置是否测试环境
	Bid      int           // ckv bid, tlist bid
	Cid      int           // tlist cid
	Db       int           // redis default connect db
}

var (
	once sync.Once
	conf = make(map[string]*Conf)
)

// Request Requestor接口的具体实现，包括业务失败，逻辑失败，attr/jm监控
type Request struct {
	ReqType int           // request type: SendAndRecv SendAndRecvKeepalive SendOnlyKeepalive SendOnly SendAndRecvIgnoreError
	Network string        // tcp udp unix zmq
	Address string        // l5://modid:cmdid  ip://ip:port  cmlb://appid  dns://id.qq.com:80  sock://filepath
	Timeout time.Duration // current action timeout time.Second
	Sidecar bool          // if true, address redirect to 127.0.0.1:65001

	ErrCode int           // return error code after finish
	IPPort  string        // return ip:port address after addressing
	Cost    time.Duration // return cost time after finish

	Command        string // service request command name, for jm report
	Prefix         string // for jm report
	Sequence       uint32 // service packet sequence
	ServiceErrCode int    // for monitor
	ServiceErrMsg  string // for monitor

	cl5Addr string // cl5寻址时，单独维护一份，防止一个req多次SetCl5HashKey
}

// GetCommuErrCode get commu error code
func (r *Request) GetCommuErrCode() int {
	return r.ErrCode
}

// GetCommuErrMsg get commu error message
func (r *Request) GetCommuErrMsg() string {
	return ErrMsg[r.ErrCode]
}

// GetServiceErrCode get service error code
func (r *Request) GetServiceErrCode() int {
	return r.ServiceErrCode
}

// GetServiceErrMsg get service error message
func (r *Request) GetServiceErrMsg() string {
	return r.ServiceErrMsg
}

// GetErrCode get final error code
func (r *Request) GetErrCode() int {
	if r.ErrCode != 0 {
		return r.ErrCode
	} else if r.ServiceErrCode != 0 {
		return r.ServiceErrCode
	}
	return 0
}

// DataSourceName data source name  cmlb://39990?timeout=800&reqtype=0&network=udp
func (r *Request) DataSourceName() string {
	if r.Sidecar == true {
		return fmt.Sprintf("%s?timeout=%d&reqtype=%d&network=%s", "ip://127.0.0.1:65001", r.Timeout/time.Millisecond, r.ReqType, r.Network)
	}
	return fmt.Sprintf("%s?timeout=%d&reqtype=%d&network=%s", r.Address, r.Timeout/time.Millisecond, r.ReqType, r.Network)
}

// Cmd command
func (r *Request) Cmd() string {
	return r.Command
}

// Check returns data length by default
func (r *Request) Check(data []byte) (int, error) {
	return len(data), nil
}

// Marshal returns empty by default
func (r *Request) Marshal() ([]byte, error) {
	return nil, nil
}

// Unmarshal returns success by default
func (r *Request) Unmarshal(data []byte) error {
	return nil
}

// Finish set error code, address, cost time when request finish
func (r *Request) Finish(ec int, addr string, cost time.Duration) {
}

// Success check if error code is ok
func (r *Request) Success() bool {
	return r.ErrCode == 0 && r.ServiceErrCode == 0
}

// IsTimeout check if error code is timeout
func (r *Request) IsTimeout() bool {
	return r.ErrCode == ErrRecvTimeout || r.ErrCode == ErrContextTimeout
}

// IsCanceled check if error code is context canceled
func (r *Request) IsCanceled() bool {
	return r.ErrCode == ErrContextCanceled
}

// SetTimeout set timeout like time.Second or 800 * time.Millisecond
func (r *Request) SetTimeout(d time.Duration) {
	if d == 0 {
		return
	}
	r.Timeout = d
}

// SetAddress set address like ip://ip:port or l5://modid:cmdid
func (r *Request) SetAddress(s string) {
	if s == "" {
		return
	}
	r.Address = s
}

// SetNetwork set network: tcp udp unix
func (r *Request) SetNetwork(s string) {
	if s == "" {
		return
	}
	r.Network = s
}

// SetReqType set request type:SendAndRecv SendAndRecvKeepalive	SendOnlyKeepalive SendOnly SendAndRecvIgnoreError
func (r *Request) SetReqType(t int) {
	if t == 0 {
		return
	}
	r.ReqType = t
}

// SetCl5HashKey 设置cl5一致性哈希的key，配置address必须是 cl5://modid:cmdid
func (r *Request) SetCl5HashKey(k uint64) {
	parts := strings.Split(r.Address, ":")
	if len(parts) == 0 {
		return
	}
	// 同时兼容cl5打头或者l5://mod:cmd:key
	if parts[0] == "cl5" || (parts[0] == "l5" && len(parts) == 4) {
		if r.cl5Addr == "" {
			r.cl5Addr = r.Address
		}
		r.Address = fmt.Sprintf("%s:%d", r.cl5Addr, k)
	}
}

// String return err msg with ec addr cost
func (r *Request) String() string {
	return fmt.Sprintf("%s, request[%s], cost[%s], addr[%s], result[%d,%s]", r.Command, ErrMsg[r.ErrCode], r.Cost.String(), r.IPPort, r.ServiceErrCode, r.ServiceErrMsg)
}

// DebugString return err msg with full request info
func (r *Request) DebugString() string {
	return fmt.Sprintf("request[%s], cost[%s], request info:%#v", ErrMsg[r.ErrCode], r.Cost.String(), r)
}

func (r *Request) ErrorString() string {
	return fmt.Sprintf("[NetErr %d:%s] [Addr %s] [Cost %s] [Net %s] [ReqType %s]",
		r.ErrCode,
		ErrMsg[r.ErrCode],
		r.IPPort,
		r.Cost.String(),
		r.Network,
		ReqTypeMsg[r.ReqType])
}
