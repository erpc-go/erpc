package server

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// HandlerFunc 定义命令字对应处理函数类型
type HandlerFunc func(ctx *Context)

// Process 处理handler函数定义接口
func (f HandlerFunc) Process(ctx *Context) {
	f(ctx)
}

// Handler 处理handler接口定义
type Handler interface {
	Process(ctx *Context)
}

func GetPackageName() (string, error) {
	pwd, _ := os.Getwd()
	s := strings.Split(pwd, "/")
	if len(s) < 5 {
		return "", fmt.Errorf("invalid pwd:%s", pwd)
	}
	if s[1] != "usr" || s[2] != "local" || s[3] != "services" {
		return "", fmt.Errorf("invalid pwd:%s", pwd)
	}
	index := strings.LastIndex(s[4], "-")
	if index == -1 {
		return "", fmt.Errorf("invalid pwd:%s", pwd)
	}
	return s[4][0:index], nil
}

// func serviceRegister(stop <-chan struct{}, registeredServices []*Service, mapEntries map[string]mutexEntry, succ chan *Service) {
// 	log.Raw("start to register.\n")
// 	// 重置环境变量
// 	var env string
// 	for _, v := range registeredServices {
// 		env += v.Marshal() + ";"
// 	}

// 	retry := make(chan mutexEntry, len(mapEntries))
// 	defer func() {
// 		close(retry)
// 	}()
// 	for _, v := range mapEntries {
// 		if len(v.token) == 0 {
// 			continue
// 		}
// 		retry <- v
// 	}
// 	packageName, err := GetPackageName()
// 	if err != nil {
// 		log.Raw("%+v\n", err)
// 	}
// 	for {
// 		select {
// 		case <-stop:
// 			log.Raw("stop service register\n")
// 			return
// 		case m := <-retry:
// 			s := &Service{Name: m.pattern, Token: m.token, Addr: ListenIP, Port: ListenPort}
// 			service, _ := register.NewService(register.ServiceConfig{Name: s.Name, Token: s.Token, Weight: 100, Addr: s.Addr, Port: uint16(s.Port), Version: version.Version(), PackageName: packageName})
// 			err := service.Register()
// 			if err != nil {
// 				log.Raw("register error, %s %s %s %d %s\n", s.Name, s.Token, s.Addr, s.Port, err)
// 				retry <- m
// 				time.Sleep(1500 * time.Millisecond)
// 			} else {
// 				env += s.Marshal() + ";"
// 				os.Setenv(ServiceEnvKey, env)
// 				log.Raw("env[%s]\n", os.Getenv(ServiceEnvKey))
// 				succ <- s
// 			}
// 		default:
// 			log.Raw("register done.\n")
// 			return
// 		}
// 	}
// }

// func serviceDeregister(input chan *Service) {
// 	log.Raw("start to deregister.\n")
// 	for {
// 		select {
// 		case s := <-input:
// 			service, _ := register.NewService(register.ServiceConfig{
// 				Name: s.Name, Token: s.Token, Addr: s.Addr, Port: uint16(s.Port),
// 			})
// 			err := service.Deregister()
// 			if err != nil {
// 				log.Raw("deregister error, %+v %s\n", s, err)
// 				input <- s
// 				time.Sleep(500 * time.Millisecond)
// 			} else {
// 				log.Raw("deregister succ, %+v\n", s)
// 			}
// 		default:
// 			close(input)
// 			log.Raw("deregister done.\n")
// 			return
// 		}
// 	}
// }

func handleSignal(succ chan *Service, stop chan<- struct{}) {
	chanSignal := make(chan os.Signal)
	signal.Notify(chanSignal, syscall.SIGTERM)
	for {
		sig := <-chanSignal
		switch sig {
		case syscall.SIGTERM:
			// 另开一个协程进行stop信号同步
			go func() {
				stop <- struct{}{}
			}()
			// 只有触发kill -TREM信号后才主动执行解注册
			// serviceDeregister(succ)
		default:
		}
	}
}

// Check 多协议包头判断
func Check(data []byte) (int, error) {
	return 0, nil
}
