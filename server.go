package erpc

import (
	"github.com/erpc-go/erpc/server"
	"github.com/erpc-go/jce-codec"
)

// GlobalServerMutex 全局唯一入口配置
var GlobalServeMutex server.ServeMutex

// HandlerFunc 注册处理函数，函数形式
func HandleFunc(pattern, token string, handler func(*server.Context), reqType, rspType jce.Messager) {
	GlobalServeMutex.HandleFunc(pattern, token, server.HandlerFunc(handler), reqType, rspType)
}

// Handle 注册处理函数，接口形式
func Handle(pattern, token string, handler server.Handler, reqType, rspType jce.Messager) {
	GlobalServeMutex.Handle(pattern, token, handler, reqType, rspType)
}

// Alias 服务别名绑定，将src服务名对应的func依次绑定到dst对应的服务名上
func Alias(src string, dst ...string) {
	GlobalServeMutex.Alias(src, dst...)
}

func ListenAndServe() {
	GlobalServeMutex.Listen()
}
