package client

import (
	"fmt"
	"time"
)

// Node 服务节点信息
type Node struct {
	ServiceName string
	Address     string // ip:port
	Network     string // tcp udp
	CostTime    time.Duration
	Metadata    map[string]interface{}
}

func (n *Node) String() string {
	return fmt.Sprintf("service:%s, addr:%s, cost:%s", n.ServiceName, n.Address, n.CostTime)
}

func (n *Node) Addr() string {
	return n.Address
}
