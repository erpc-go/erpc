## erpc
一个简易、美丽的 rpc 框架


## 抽象说明

```
客户端                  服务端
client     <---->       server

     ----   request  --- >
     <---   response ---


一次 request+response = context

而每次发送的 request 或 response 都分为 header、body

```

## 主要目录说明：
- client: 客户端实现 
- server: 服务端实现
- center：注册中心实现
- codec:  编码方式实现
- protocol: 网络协议设计
- transport: 底层网络支撑，包括连接池，连接事务等


## todo
- [x] 注册中心
- [x] 多种编码方式
- [x] 网络协议设计
- [ ] 支持多种负载均衡算法
- [ ] 把负载均衡从 center 拆分出来单独作为一个新的服务
- [ ] 增加超时处理
- [ ] 实现连接池
- [ ] 实现链路追踪日志管理
- [ ] 自定义序列化协议实现

## 现有问题
- server listen 比较卡
- center ping 有 bug
