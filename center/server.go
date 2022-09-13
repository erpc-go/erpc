package center

import (
	"math/rand"
	"time"

	"github.com/edte/erpc/log"
)

// FIX: 这里有个问题：如果部署不同机子是同一 server，但是 funcs 不同，则会出问题
// 现有解决办法是：只允许部署的 server funcs 相同（话说肯定要相同吧，都是一套代码）
// rpc 服务
// server 下面有很多 funcs
// 而每台 server 都有很多 addr，可以部署在每个 addr 上
type serverItem struct {
	name     string          // servername
	addrList []string        // server addr list
	addrMap  map[string]bool // judge addr has registe
	funcList []string        // funcname list
	funcMap  map[string]bool // judge funcname has registe
}

func newServerItem(name string) *serverItem {
	s := &serverItem{
		name:     name,
		addrList: []string{},
		addrMap:  map[string]bool{},
		funcList: []string{},
		funcMap:  map[string]bool{},
	}
	return s
}

func (s *serverItem) registe(addr string, funcs []string) {
	log.Debugf("serve %s begin registe addr, addr:%s,funcs:%v", s.name, addr, funcs)

	// [step 1] 如果地址没有注册，则注册地址
	if !s.addrMap[addr] {
		s.addrMap[addr] = true
		s.addrList = append(s.addrList, addr)
	}

	log.Debugf("serve %s begin registe funcs", s.name)

	// [step 2] 注册函数，扫描然后判断是否 func 已经注册过
	for _, f := range funcs {
		if !s.funcMap[f] {
			s.funcMap[f] = true
			s.funcList = append(s.funcList, f)
		}
	}

	log.Debugf("serve %s registe succ,data:%v", s.name, s)
}

func (s *serverItem) emptyFuncs() bool {
	return len(s.funcList) == 0
}

func (s *serverItem) emptyIp() bool {
	return len(s.addrList) == 0
}

// TODO: 负载均衡这里需要扩展，暂时随机返回一个即可
func (s *serverItem) banlance() (addr string, err error) {
	log.Debugf("begin select a ip by banlance,addrs:%v", s.addrList)

	r := rand.New(rand.NewSource(time.Now().Unix()))
	i := r.Intn(len(s.addrList))

	log.Debugf("select ip succ, res:%s", s.addrList[i])

	return s.addrList[i], nil
}
