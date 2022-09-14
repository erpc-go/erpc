package center

import (
	"errors"
	"fmt"
	"time"

	"github.com/edte/erpc/log"
)

// FIX: 这里有个问题：如果部署不同机子是同一 server，但是 funcs 不同，则会出问题
// 现有解决办法是：只允许部署的 server funcs 相同（话说肯定要相同吧，都是一套代码）
// rpc 服务
// server 下面有很多 funcs
// 而每台 server 都有很多 addr，可以部署在每个 addr 上
type serverItem struct {
	// server name
	name string

	// addrs map, 一共三种状态：1. ！ok 表示未注册， 2.true 表示有效， 3.false 表示无效
	addrs addrs

	// judge funcname has registe
	funcs funcs
}

func newServerItem(name string) *serverItem {
	s := &serverItem{
		name:  name,
		addrs: map[string]*addrItem{},
		funcs: map[string]bool{},
	}
	return s
}

// 如果 map 有该 addr，说明注册过，与是否有效无关
func (s *serverItem) hasRegiste(addr string) bool {
	_, ok := s.addrs[addr]
	return ok
}

func (s *serverItem) registe(addr string, funcs []string) {
	log.Debugf("serve %s begin registe addr, addr:%s,funcs:%v", s.name, addr, funcs)

	// [step 1] 如果地址没有注册，则注册地址
	if !s.hasRegiste(addr) {
		s.addrs.addAddr(addr)
	}

	log.Debugf("serve %s begin registe funcs", s.name)

	// [step 2] 注册函数，扫描然后判断是否 func 已经注册过
	s.funcs.addFunc(funcs)

	log.Debugf("serve %s registe succ,data:%v", s.name, s)
}

func (s *serverItem) emptyFuncs() bool {
	return s.funcs.empty()
}

func (s *serverItem) emptyIp() bool {
	return s.addrs.empty()
}

// 负载均衡选择
func (s *serverItem) banlance() (addr string, err error) {
	return s.addrs.balance()
}

// 更新无效 addr
func (s *serverItem) update() {
	s.addrs.update(s.name)
}

// 更新心跳
func (s *serverItem) heatbeat(addr string, last int64) {
	s.addrs.heatbeat(addr, last)
}

func (s *serverItem) String() string {
	return fmt.Sprintf("%v", *s)
}

type servers map[string]*serverItem

func (s servers) hasRegiste(server string) bool {
	_, ok := s[server]
	return ok
}

func (s servers) registe(server string, addr string, funcs []string) (err error) {
	log.Debugf("begin registe sever:%s", server)

	// [step 1] 如果 server 没有注册过,则在 center 中注册 server
	if !s.hasRegiste(server) {
		log.Debugf("server %s has not registerd, begin regise to center", server)
		s[server] = newServerItem(server)
	}

	// [step 2] 如果 server 已经注册过
	log.Debugf("server %s begin regise funcs,funcs:%v", server, funcs)
	s[server].registe(addr, funcs)

	log.Debugf("server %s registe succ", server)

	return

}

func (s servers) discovery(server string) (addr string, err error) {
	log.Debugf("begin get server list %s", server)

	// [step 1] 先从 server map 里取 server， 如果不存在则返回
	if !s.hasRegiste(server) {
		log.Errorf("serve %s discover failed, err:%s", server, "server not register")
		return "", errors.New("server not register")
	}

	log.Debugf("begin check server %s", server)

	// [step 2] 如果 server 部署的服务器为空，则返回
	if s[server].emptyIp() {
		log.Errorf("server %s discover failed, ip is empty", server)
		return "", errors.New("server's address list is empty")
	}

	// [step 3] 如果 server 部署的服务器 func 为空，则返回
	if s[server].emptyFuncs() {
		log.Errorf("server %s discover failed, funcs is empty", server)
		return "", errors.New("server's funcs list is empty")
	}

	log.Debugf("begin banlance server %s addr", server)

	// [step 3] 正常则帅负载均衡算法选一个服务器
	return s[server].banlance()
}

func (s servers) update() {
	for {
		log.Debugf("begin update server")

		t := time.NewTicker(time.Second)
		select {
		case <-t.C:
			for _, v := range s {
				v.update()
			}

			log.Debugf("update server succ")
		}
	}
}

func (s servers) String() (res string) {
	for k, v := range s {
		res += fmt.Sprintf("%s : %s", k, v)
	}

	return
}
