参考 net、gin 的设计

todo:
1. 当 server handler 中发生了 panic 之后，整个服务都会 down 掉，需要加 recover 处理


# `tme-protocol` 服务端框架

- 支持多种自定义协议，如`pdu`、`qza`、`tme`、`HTTP`
- 自动进行协议探测
- 支持服务热重启
- 支持服务注册、解注册

## 1. 服务端注册模式

- HandleFunc/Handle仅支持微服务命令字的注册
- qza/pdu/HTTP非微服务需使用Alias()方法进行绑定, 且需绑定到微服务命令字上

```golang
func main() {
	// 注册微服务协议handler -- handleA
	ths.HandleFunc("tme.microservices.test.1", "5ae27290-ba3b-4b33-b579-2df16bf09b71", handleA, (*mini_game_sdk.SdkInputSvrReq)(nil), (*mini_game_sdk.SdkInputSvrRsp)(nil))

	// 绑定qza协议 -- 与微服务共用同一个handlerA
	ths.Alias("tme.microservices.test.1", "/qza0x80/4")

	// 绑定qza协议 -- 与微服务共用同一个handlerA
	ths.Alias("tme.microservices.test.1", "/pdu0x0/4")

	// 绑定HTTP协议 -- 与微服务共用同一个handlerA
	ths.Alias("tme.microservices.test.1", "/rif/room/room_list_server/get_list")

	// 服务启动
	ths.ListenAndServe()
}

func handleA(ctx *ths.Context) {
	// todo
}

```

## 2. 相关概念解释

- server：代表一个服务实例，即一个进程。（一般对应[TME运维平台](http://music.isd.com)上的一个包）
- service：代表一个接口，一个server可以包含多个service。
  - 微服务：即对应一个服务名，如`kge.rec.live.push`
  - qza、pdu：由协议、cmd、subcmd组成的三元组，如`/qza0x80/4`，`/pdu0x0/0`
  - HTTP: 由HTTP Request Path组成
- token：服务注册鉴权标识。
  - 微服务：从[服务治理平台](http://kp.kg.oa.com/projects/microservice/#/service-manager/service-list)获取，**作为`HandleFunc()`第2个参数传递**
  - qza、pdu、HTTP：无需配置, 使用Alias方法进行handler绑定
- handler：即处理函数。一般一个server对应一个hanlder，当然也可以多个service对应同一个handler

## 3、配置

- ths使用`going/config`库解析配置文件，默认路径 `../conf/config.toml`。业务如需更换配置文件路径，需在调用ListenAndServe()前显式赋值`going/config`包级变量`ConfPath`
- 配置struct为`going/cat`包的[ServerInfo](https://git.woa.com/tme/going/blob/master/cat/config.go)
- 配置样例:

```yaml
[server]
name = "ths-test"
user = "test"
addr = "eth1:6868/eth1"
msgTimeout = "800ms"
enableDebugMode = true
```

## 4. 服务端执行流程

1. accept一个新链接启动一个goroutine接收该链接数据
2. 收到一个完整数据包，解包整个请求
3. 查询handler map，定位到具体处理函数
4. 反序列化请求body
5. 调用业务处理函数
6. 序列化响应body
7. 打包整个响应
8. 回包给上游客户端

最近更新于 2021.09.16
