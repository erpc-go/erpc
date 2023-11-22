package server

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	ServiceEnvKey = "SERVICES"
)

type Service struct {
	Name  string
	Token string
	Addr  string
	Port  int
}

func NewServivce(s string) (*Service, error) {
	service := strings.Split(s, ",")
	if 4 != len(service) {
		return nil, errors.New(fmt.Sprintf("Service Unmarshal failed, len %d != 4 [%s]", len(service), s))
	}
	result := &Service{
		Name:  service[0],
		Token: service[1], Addr: service[2],
	}
	result.Port, _ = strconv.Atoi(service[3])
	return result, nil
}

func (s *Service) Marshal() string {
	return fmt.Sprintf("%s,%s,%s,%d", s.Name, s.Token, s.Addr, s.Port)
}

func ParseServiceFromEnv() []*Service {
	var result []*Service
	for _, env := range os.Environ() {
		// 遍历环境变量
		s := strings.Split(env, "=")
		if 2 != len(s) {
			continue
		}
		// 解析service列表
		if s[0] == ServiceEnvKey {
			ss := strings.Split(s[1], ";")
			for i := range ss {
				if len(ss[i]) > 0 {
					// 解析service元信息
					service, err := NewServivce(ss[i])
					if err != nil {
						continue
					}
					result = append(result, service)
				}
			}
		}
	}
	return result
}
