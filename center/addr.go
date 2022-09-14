package center

import (
	"math/rand"
	"time"

	"github.com/edte/erpc/log"
)

type addrs map[string]*addrItem

type addrItem struct {
	addr     string // ip:port
	valid    bool   // if addr valid
	lasttime int64  // last update time
}

func (a addrs) addAddr(addr string) {
	a[addr] = &addrItem{
		addr:     addr,
		valid:    true,
		lasttime: time.Now().UnixMilli(),
	}
}

// judge if addrs valid empty
func (a addrs) empty() bool {
	for i := range a {
		if a[i].valid {
			return true
		}
	}
	return false
}

// TODO: 负载均衡这里需要扩展，暂时随机返回一个即可
func (a addrs) balance() (addr string, err error) {
	tmp := []string{}

	for _, v := range a {
		if v.valid {
			tmp = append(tmp, v.addr)
		}
	}

	r := rand.New(rand.NewSource(time.Now().Unix()))
	i := r.Intn(len(tmp))

	return tmp[i], nil
}

func (a addrs) update(server string) {
	now := time.Now()

	log.Debugf("begin update addr, raw:%v", a)

	// [setp 1] 扫描 addr
	for _, v := range a {
		// [step 2] 如果 addr 无效，则不用更新
		if !v.valid {
			continue
		}

		// [step 3] 计算上次的心跳时间差
		lasttime := time.UnixMilli(v.lasttime)
		d := now.Sub(lasttime)

		log.Debugf("addr %s lasttime:%s, now:%s, distance %s, ", v.addr, lasttime, now, d)

		// [step 4] 如果小于 5s，则跳过
		if d <= time.Second*5 {
			continue
		}

		log.Debugf("server %s's addr %s begin invalid", server, v.addr)

		// [step 5] 大于 5s，则更新为无效 addr，同时从有效 list 中删除
		v.valid = false
	}
}

// 更新心跳
func (a addrs) heatbeat(addr string, last int64) {
	// [step 1] 更新上次心跳时间
	a[addr].lasttime = last

	// [step 2] 让 addr 有效
	a[addr].valid = true
}
