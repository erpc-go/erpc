网络底层实现，包括

- 连接上下文封装
- 连接池
- 网络模型

其中，上下文是一次 request-response 的封装，连接池则主要用于 client 端维护，而网络模型则是自定义的一个 reactor 模型，代替 net 自带的 poll 模型
