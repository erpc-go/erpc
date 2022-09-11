package center

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/edte/erpc/log"
	ping2 "github.com/go-ping/ping"
)

// TODO: 暂时注册中心写在本地，之后迁移到服务上

// server:funcs  -> 集群

// 一个 server 下面有很多 func, 是逻辑独立的
// 而每个 server 都是一个集群，每个 server 可以部署在一台机器上
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

type serverItem struct {
	name      string               // servername
	addr      []string             // server ip addr
	funcsList []*funcItem          // func list
	funcs     map[string]*funcItem // funcName -> func info
}

func newServerItem(name string, addr string, funcName string, i *funcItem) *serverItem {
	s := &serverItem{
		name:      name,
		addr:      []string{addr},
		funcsList: []*funcItem{i},
		funcs:     map[string]*funcItem{funcName: i},
	}
	return s
}

type funcItem struct {
	name     string      // func name
	request  interface{} // request type
	response interface{} // response type
}

func newFuncItem(name string, req interface{}, rsp interface{}) *funcItem {
	f := &funcItem{
		name:     name,
		request:  req,
		response: rsp,
	}

	return f
}

func (c *Center) register(server string, addr string, request interface{}, response interface{}) (err error) {
	log.Debugf("begin split server, server:%s", server)

	// [step 1]  分割 server，格式为 servername.funcname,如果不满足则失败
	s := strings.Split(server, ".")
	if len(s) != 2 {
		return errors.New("invalid server:" + server)
	}

	// alias 下，方便编写
	serverName := s[0]
	funcName := s[1]

	log.Debugf("split server %s succ, servername:%s, funcName:%s", serverName, funcName)

	// [step 2] 建立 func item
	fi := newFuncItem(funcName, request, response)

	log.Debugf("begin judge server %s has registe", serverName)

	// [step 3] 如果 server 已经注册过
	if oldServer, ok := c.serverInfo[serverName]; ok {
		log.Debugf("server %s has registerd", serverName)
		log.Debugf("begin regise serve %s's func %s", serverName, funcName)

		// [step 3.1] 增加 server 地址
		oldServer.addr = append(oldServer.addr, addr)

		// [step 3.2] 然后判断是否该 server 已注册该函数,如果没，则增加
		if _, ok = oldServer.funcs[funcName]; !ok {
			oldServer.funcsList = append(oldServer.funcsList, fi)
			oldServer.funcs[funcName] = fi
		}

		log.Debugf("server %s registe succ", server)

		return
	}

	log.Debugf("begin registe server %s", server)

	// [step 4] 如果 server 没有注册过,则注册 server
	si := newServerItem(serverName, addr, funcName, fi)

	c.serverList = append(c.serverList, si)
	c.serverInfo[serverName] = si

	log.Debugf("server %s registe succ", server)

	return nil
}

func (c *Center) discovery(server string) (addr string, err error) {
	log.Debugf("begin discovey server %s", server)

	// [step 1] 先从 server map 里取 server， 如果不存在则返回
	si, ok := c.serverInfo[server]
	if !ok {
		log.Errorf("serve %s discover failed, err:%s", server, "server not register")
		return "", errors.New("server not register")
	}

	// [step 2] 如果 server 部署的服务器为空，则返回
	if len(si.addr) == 0 {
		log.Errorf("server %s discover failed, ip is empty", server)
		return "", errors.New("server's address list is empty")
	}

	// [step 3] 正常则帅负载均衡算法选一个服务器
	return c.banlance(si.addr)
}

func (c *Center) banlance(addrs []string) (addr string, err error) {
	// TODO: 负载均衡这里需要扩展，暂时随机返回一个即可

	log.Debugf("begin select a ip by banlance,addrs:%v", addrs)

	r := rand.New(rand.NewSource(time.Now().Unix()))
	i := r.Intn(len(addrs))

	log.Debugf("select ip succ, res:%s", addrs[i])

	return addrs[i], nil
}

// 扫描所有 server，然后 ping，剔除无效的 server
// 再扫描无效 server list，如果没问题，则放入正常 server list
// TODO: ping 命令这里暂时用库，以后自己实现
func ping() {
	log.Debugf("begin ping")

	// 临时的重新有效 list
	tmp := make([]*serverItem, 0)

	log.Debugf("begin scan invalidServerList")

	// [step 1] 扫描失效 server list
	for i, is := range defaultCenter.invalidServerList {
		// [step 1.1] 扫描对应 server 的集群
		for _, add := range is.addr {
			// [step 1.1.1] 发送 ping
			p, _ := ping2.NewPinger(add)
			_ = p.Run()

			// [setp 1.1.2] 如果该机器响应在 1s 内，则说明可以成功响应, 则放回有效 list 中, 并且从无效 list 中删除
			// 超出则说明还是无效，无视即可
			statis := p.Statistics()
			if statis.AvgRtt <= time.Second {
				tmp = append(tmp, is)
				defaultCenter.invalidServerList = append(defaultCenter.invalidServerList[:i], defaultCenter.serverList[i+1:]...)
			}
		}
	}

	log.Debugf("begin scan server list")

	// [step 2] 扫描有效 server list
	for i, is := range defaultCenter.serverList {
		// [step 2.1] 扫描对应 server 的集群
		for _, add := range is.addr {
			// [step 1.1.1] 发送 ping
			p, _ := ping2.NewPinger(add)
			_ = p.Run()

			// [setp 1.1.2] 如果该机器响应大于 1s，则说明该 server 无效
			// 则从有效 list 中删除，然后加入到无效 list 中
			statis := p.Statistics()
			if statis.AvgRtt > time.Second {
				defaultCenter.serverList = append(defaultCenter.serverList[:i], defaultCenter.serverList[i+1:]...)
				delete(defaultCenter.serverInfo, is.name)
				defaultCenter.invalidServerList = append(defaultCenter.invalidServerList, is)
			}
		}
	}

	log.Debugf("begin generate new server list, invalidServerList:%v", tmp)

	// [step 3] 把重新有效的 server 放入到有效 list 中
	for _, is := range tmp {
		defaultCenter.serverList = append(defaultCenter.serverList, is)
		defaultCenter.serverInfo[is.name] = is
	}

	log.Debugf("ping finished")
}

func Ping() {
	// 周期 ping server
	for {
		t := time.NewTicker(time.Second)
		select {
		case <-t.C:
			ping()
		}
	}
}
