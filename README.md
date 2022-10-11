## erpc
一个简易、优雅的 rpc 框架


## 安装
```
go get github.com/edte/erpc
```

## 抽象

```
客户端                  服务端
client     <---->       server

     ----   request  --- >
     <---   response ---


一次 request+response = context

而每次发送的 request 或 response 都分为 header、body

```

## 主要目录
- client: 客户端实现 
- server: 服务端实现
- center：注册中心实现
- codec:  编码方式实现
- protocol: 网络协议设计
- transport: 底层网络支撑，包括连接池，连接事务等


## 使用
**过程**

1. 创建服务传输协议（默认 pb），可使用 webhook 自动部署到 github 
2. 部署注册中心 center
3. 部署服务模块 server
4. 部署客户端模块 client

**例子**

```go
go run ./demo/center/center.go
go run ./demo/server/server.go
go run ./demo/client/client.go
```


## 快速开始


**center**
```go
func main() {
	erpc.ListenCenter()
}
```

**server**

```go

func handleHello(c *transport.Context) {
	rsp := c.Response.(*demo.HelloResponse)

	rsp.Msg = "hello world"
	fmt.Println(rsp.Msg)
}

func main() {
	erpc.Handle("demo.hello", handleHello, &demo.HelloRequest{}, &demo.HelloResponse{})
	erpc.Listen(":8877")
}

```

**client**
```go
func main() {
    req := demo.HelloRequest{}
    rsp := demo.HelloResponse{}
    
    ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
    
    if err := erpc.Call(ctx, "demo.hello", &req, &rsp); err != nil {
    	panic(fmt.Sprintf("call demo.hello failed, error:%s", err))
    }

    fmt.Println(rsp.Msg)
}
```

## 序列化方式
支持多种序列化方式，提供：
1. jce(默认)
2. pb
3. Avro
4. json
5. MessagePack
6. Thrift
7. XML
8. SOAP
9. None(不编码，要求数据本身就为 []byte)
10. 自定义

## 网络协议
1. tcp(默认)
2. udp
3. quic
4. unix socket
5. kcp
6. http
7. websocket


## 注册中心
1. eregister(默认)
2. 点对点
3. 点对多
4. zookeeper
5. etcd
6. consul
7. mDNS
8. nacos
9. 本进程(调试)



## 失败重试算法
1. failfast(不重试)
2. failover（
3. failtry
4. failbackup

## 路由算法
或者说负载均衡算法

1. 随机
2. 轮询
3. 一致性 hash
4. 权重
5. 网络质量优先
6. 地理位置优先
7. 自定义


## 压缩算法

1. gzip
2. hufuman
3. zlib
4. none
5.

## 限流算法
1. 漏桶
2. 令牌桶


## 熔断机制


## 连接池

## 协程池

## 网络模型



## 身份认证



## API 网关


## 全双工通信


## Metrics
服务统计


## Mock 机制
测试

## 配置中心


## 链路日志


## 协议设计
1. request 和 response 还是 message？（格式一个还是两个）
2. header 中的 map 怎么设计（是原生编码，还是 jce？）
3. response code 怎么设计？
4. request 和  response header 中都设计 servername、funcname?
5. servername 和 funcname 要分开吗？（两部分还是三部分）

有哪些需要设计的？
有哪些不应该设计？
有哪些字段？
扩展性呢？
版本迭代怎么做？



## todo
- [x] 序列化协议设计
- [ ] 网络协议设计
- [x] 客户端实现
- [x] 服务端实现
- [x] 支持多种编码方式
- [x] 增加超时处理
- [x] 心跳检测功能
- [ ] 支持多种负载均衡算法
- [ ] 注册中心实现
- [ ] 实现连接池
- [ ] 实现链路追踪日志管理
- [ ] 支持多种失败重试机制
- [ ] 自定义序列化协议实现
- [ ] 鉴权机制
- [ ] 同步、异步客户端实现
- [ ] mock 测试实现
- [ ] 限流实现
- [ ] 熔断机制
- [ ] 自定义网络模型
- [ ] 协程池实现
- [ ] 把负载均衡从 center 拆分出来单独作为一个新的服务
- [ ] center 支持路由功能
- [ ] center 支持多机房部署

## 现有问题
1. server 死循环
client 断开之后，server 之前 accept 的那个协程会一直轮询，考虑优化掉(方案1：解码编码处优化，方案2：线程池优化，考虑关闭 client 后就自动) sever 就关闭该协程

2. 注册中心地址
怎么知道注册中心地址，这是一个问题，也就是说，server 向 center 请求其他 server 的地址，前提是知道  center 的地址在哪，这个要看怎么实现，不可能 ip 写死到代码中，因为 center 也是一个集群，本项目暂时吧 center 作为基础服务，每个 server 部署的机器上都有一个 center，所以固定 center 的监听端口，然后 server 请求时向这个默认端口请求即可。
但是问题也有很多，比如每台机器上都部署 center 这种方式好吗？并且，center 之间的同步怎么做？（考虑加 redis 缓存？感觉不太好，还是一般的分布式同步算法？待重新设计）
