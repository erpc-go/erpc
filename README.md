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

1. 使用序列化协议（默认 pb） 创建传输 struct
2. 部署 center
3. 部署 server
4. 部署 client

**例子**

```
go run ./demo/center/center.go
go run ./demo/server/server.go
go run ./demo/client/client.go
go run ./demo/balance/balance.go
```

## todo
- [x] 注册中心
- [x] 多种编码方式
- [x] 网络协议设计
- [x] 增加超时处理
- [x] 心跳检测
- [ ] 支持多种负载均衡算法
- [ ] 把负载均衡从 center 拆分出来单独作为一个新的服务
- [ ] 实现连接池
- [ ] 实现链路追踪日志管理
- [ ] 自定义序列化协议实现
- [ ] 熔断机制
- [ ] 自定义网络模型
- [ ] 协程池实现

## 现有问题
- center ping 有 bug
- 处理延时过大
