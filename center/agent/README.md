注册中心代理，部署在每台机器上，作为基础服务。

提供以下服务：
1. 缓存本地 server -> iplist
2. 提供客户端接口，即 server 得到 iplist
3. 提供负载均衡算法，实现 server -> ip
4. 提供服务订阅接口，自动更新 list

