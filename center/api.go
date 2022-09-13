package center

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/edte/erpc/log"
	"github.com/gin-gonic/gin"
)

var (
	defaultCenter = NewCenter()
)

type ResigerArg struct {
	Server    string
	Addr      string
	Functions []string
}

// TODO: 这里同时提供 http、rpc 两种与服务中心交付的方式（哪种兜底？）
// TODO: 再考虑本地服务缓存提供服务发现、服务注册功能？（去注册中心化？考虑优化）

// 服务注册
// 一台机器向注册中心发送本地地址、注册的服务和接口，以及接口的参数、返回值
func Register(arg ResigerArg) (err error) {
	b, err := json.Marshal(arg)
	if err != nil {
		return
	}

	resp, err := http.Post("http://127.0.0.1:8080/register", "application/json", strings.NewReader(string(b)))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode == 500 {
		return errors.New(string(body))
	}

	return
}

// 服务发现
// 给定请求服务名，然后负债均衡返回其中一个部署的机器 ip 地址
// TODO: 每次响应一个 ip，那其他集群内怎么同步的？
func Discovery(server string) (addr string, err error) {
	resp, err := http.Get("http://127.0.0.1:8080/discovery?server=" + server)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode == 500 {
		return "", errors.New(string(body))
	}

	return string(body), nil
}

func Listen(addr string) {
	r := gin.Default()

	r.GET("/discovery", func(ctx *gin.Context) {
		s := ctx.Query("server")
		res, err := defaultCenter.discovery(s)
		if err != nil {
			log.Errorf("server %s discovery failed, err:%s", s, err)
			ctx.String(500, err.Error())
			return
		}
		log.Debugf("serve %s discover succ, res:%s", s, res)
		ctx.String(200, res)
	})

	r.POST("/register", func(ctx *gin.Context) {
		r := &ResigerArg{}

		if err := ctx.BindJSON(&r); err != nil {
			log.Errorf("registe failed, err:%s", err)
			ctx.String(500, err.Error())
			return
		}

		if err := defaultCenter.register(r.Server, r.Addr, r.Functions); err != nil {
			log.Errorf("registe failed, err:%s", err)
			ctx.String(500, err.Error())
			return
		}

		log.Debugf("register %s succ", r.Server)
	})

	go BeginPing()

	r.Run()
}
