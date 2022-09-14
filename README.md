## erpc
一个简易、优雅的 rpc 框架


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


## Demo
**过程**

1. 使用序列化协议（默认 pb） 创建传输 struct(部署到 github 上)
2. 部署 center
3. 部署 server
4. 部署 client

**例子**

```
go run ./demo/center/center.go
go run ./demo/server/server.go
go run ./demo/client/client.go
```

## todo
- [x] 客户端实现
- [x] 服务端实现
- [x] 注册中心实现
- [x] 支持多种编码方式
- [x] 网络协议设计
- [x] 增加超时处理
- [x] 心跳检测功能
- [ ] 支持多种负载均衡算法
- [ ] 实现连接池
- [ ] 实现链路追踪日志管理
- [ ] 自定义序列化协议实现
- [ ] 熔断机制
- [ ] 自定义网络模型
- [ ] 协程池实现
- [ ] 把负载均衡从 center 拆分出来单独作为一个新的服务
- [ ] 支持路由

## 现有问题
client 断开之后，server 之前 accept 的那个协程会一直轮询，考虑优化掉(方案1：解码编码处优化，方案2：线程池优化，考虑关闭 client 后就自动) sever 就关闭该协程
