package client

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
)

var (
	poolMap  = make(map[string]*Pool)
	poolLock sync.RWMutex
)

func init() {
	go func() {
		for {
			time.Sleep(time.Minute)
			// var poolNum int
			var connNum int
			poolLock.RLock()
			// poolNum = len(poolMap)
			for _, p := range poolMap {
				connNum += p.Len()
			}
			poolLock.RUnlock()
		}
	}()
}

// GetTCPConnectionPool 获取tcp连接池
func GetTCPConnectionPool(key string, addr string, network string, timeout time.Duration) *Pool {
	var pool *Pool
	var ok bool

	poolLock.RLock()
	pool, ok = poolMap[key]
	poolLock.RUnlock()
	if ok {
		return pool
	}

	poolLock.Lock()
	defer poolLock.Unlock()

	pool, ok = poolMap[key]
	if ok {
		return pool
	}

	if timeout < time.Millisecond*300 {
		timeout = time.Millisecond * 300 // 防止timeout过短，每次dial必失败问题
	}
	pool, _ = NewPool(1, 10000, func() interface{} {
		c, e := net.DialTimeout(network, addr, timeout)
		if e != nil {
			log.Output(1, "dial fail:"+e.Error())
			return nil
		}
		return c
	})

	pool.Idle = 3 * time.Minute // 空闲时间3min，超过3min没有使用过的连接自动丢弃

	pool.Ping = func(conn interface{}) bool {
		if conn == nil {
			return false
		}
		if _, ok := conn.(net.Conn); !ok {
			return false
		}
		return true
	}

	pool.Close = func(conn interface{}) {
		if conn == nil {
			return
		}
		if c, ok := conn.(net.Conn); ok {
			c.Close()
		}
	}

	pool.RegisterChecker(3*time.Second, func(conn interface{}) bool {
		if conn == nil {
			return false
		}
		c, ok := conn.(net.Conn)
		if !ok {
			return false
		}

		c.SetReadDeadline(time.Now().Add(time.Millisecond))
		if _, err := c.Read(make([]byte, 1)); err == io.EOF { // 每隔3s尝试接收一个字节，探测对端是否已经主动关闭连接
			return false
		}

		return true
	})

	poolMap[key] = pool
	return pool
}
