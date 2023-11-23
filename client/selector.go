package client

import (
	"errors"
	"sort"
	"sync"
	"time"
)

var ErrNotImplement = errors.New("not implement")

type Selector interface {
	Select(serviceName string) (*Node, error)
	Update(node *Node, cost time.Duration, err error) error
}

var (
	selectors = make(map[string]Selector)
	lock      = sync.RWMutex{}
)

// Register 注册selector，如l5 dns cmlb tseer
func Register(name string, s Selector) {
	lock.Lock()
	selectors[name] = s
	lock.Unlock()
}

// Get 获取selector
func Get(name string) Selector {
	lock.RLock()
	s, ok := selectors[name]
	lock.RUnlock()
	if ok {
		return s
	} else {
		return &NoopSelector{}
	}
}

// Selectors 当前支持的selctor列表
func Selectors() []string {
	lock.RLock()
	var list []string
	for name := range selectors {
		list = append(list, name)
	}
	sort.Strings(list)
	lock.RUnlock()
	return list
}

type NoopSelector struct{}

func (noop *NoopSelector) Select(serviceName string) (*Node, error) {
	return nil, ErrNotImplement
}

func (noop *NoopSelector) Update(node *Node, cost time.Duration, err error) error {
	return ErrNotImplement
}
