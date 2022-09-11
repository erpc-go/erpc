package center

var (
	defaultCenter = NewCenter()
)

// 服务注册
// 一台机器向注册中心发送本地地址、注册的服务和接口，以及接口的参数、返回值
// TODO: 这里考虑优化：把一个 server 的所有 func 一次性请求过来注册？
// 这里注册中心主要存 req、rsp type 吗？有必要吗？
func Register(server string, addr string, request interface{}, response interface{}) (err error) {
	return defaultCenter.register(server, addr, request, response)
}

// 服务发现
// 给定请求服务名，然后负债均衡返回其中一个部署的机器 ip 地址
// TODO: 每次响应一个 ip，那其他集群内怎么同步的？
func Discovery(server string) (addr string, err error) {
	return defaultCenter.discovery(server)
}
