package center

import (
	"time"

	"github.com/edte/erpc/log"
	ping2 "github.com/go-ping/ping"
)

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
		for _, add := range is.addrList {
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
		for _, add := range is.addrList {
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

func BeginPing() {
	// 周期 ping server
	for {
		t := time.NewTicker(time.Second)
		select {
		case <-t.C:
			ping()
		}
	}
}
