package center

import (
	"errors"

	"github.com/edte/erpc/log"
)

// TODO: 把负载均衡功能拆分出来，然后服务发现的时候只返回 l5 地址，而不是直接是 ip，把选的过程拆分到 l5，让 client 去调用

// 注册中心
// 提供服务注册、服务发现功能
type Center struct {
	serverList        []*serverItem          // server list
	invalidServerList []*serverItem          // ping 没有响应的 server list
	serverInfo        map[string]*serverItem // servername -> server info
}

func NewCenter() *Center {
	c := &Center{
		serverList:        []*serverItem{},
		invalidServerList: []*serverItem{},
		serverInfo:        map[string]*serverItem{},
	}
	return c
}

func (c *Center) hasServerRegiste(server string) bool {
	_, ok := c.serverInfo[server]
	return ok
}

func (c *Center) register(server string, addr string, funcs []string) (err error) {
	log.Debugf("begin registe sever:%s", server)

	// [step 1] 如果 server 没有注册过,则在 center 中注册 server
	if !c.hasServerRegiste(server) {
		log.Debugf("server %s has not registerd, begin regise to center", server)
		si := newServerItem(server)
		c.serverList = append(c.serverList, si)
		c.serverInfo[server] = si
	}

	// [step 2] 如果 server 已经注册过
	log.Debugf("server %s begin regise funcs,funcs:%v", server, funcs)
	c.serverInfo[server].registe(addr, funcs)

	log.Debugf("server %s registe succ", server)

	return
}

func (c *Center) discovery(server string) (addr string, err error) {
	log.Debugf("begin get server list %s", server)

	// [step 1] 先从 server map 里取 server， 如果不存在则返回
	si, ok := c.serverInfo[server]
	if !ok {
		log.Errorf("serve %s discover failed, err:%s", server, "server not register")
		return "", errors.New("server not register")
	}

	log.Debugf("begin check server %s", server)

	// [step 2] 如果 server 部署的服务器为空，则返回
	if si.emptyIp() {
		log.Errorf("server %s discover failed, ip is empty", server)
		return "", errors.New("server's address list is empty")
	}

	// [step 3] 如果 server 部署的服务器 func 为空，则返回
	if si.emptyFuncs() {
		log.Errorf("server %s discover failed, funcs is empty", server)
		return "", errors.New("server's funcs list is empty")
	}

	log.Debugf("begin banlance server %s addr", server)

	// [step 3] 正常则帅负载均衡算法选一个服务器
	return si.banlance()
}
