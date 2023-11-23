rpc client 实现

参考 net/http client.go 的设计


# `tme-protocol` 客户端调用

为TMEHead接口提供了RPC Client。支持：

- 多种应用协议：如`JCE`、`Protobuf`、`HTTP`
- 多种网络模型：`TCP`（长连接、短链接）、`UDP`、`Unix Socket`
- 网络优化：连接复用、连接池

## 1、客户端调用模式

```golang
desc := thc.CallDesc{ServiceName: "tme.microservices.test.1", AppProtocol:"tme"}
req := mini_game_sdk.SdkInputSvrReq{}
rsp := mini_game_sdk.SdkInputSvrRsp{}
client := thc.New(desc, th.AuthInfo{}, &req, &rsp)
if err := client.Do(ctx); err != nil {
	ctx.Error("Client.Do failed, err_msg:%s", err)
	return
}
ctx.Debug("Client.Do succ")
```

## 2、相关概念解释

- Client：客户端调用代理。内部封装了第三方协议首部、应用层数据结构。网络层复用`going/client`统一进行连接池维护等。每次RPC新建Client即可，成本极低
- CallDesc：RPC调用参数。必须指定被调服务名`ServiceName`，可以提供`AppProtocol`明确指定协议类型：`tme`、`qza`、`pdu`。对于`qza`、`pdu`需指定`CmdID`、`SubCmdID`
- Req/Rsp：空接口类型。但目前仅支持`gojce.Message`、`proto.Message`、`*http.Response`。所以New的入参一般为实现了上述接口的**引用类型**。


## 3、配置

- thc.New使用`going/config`库解析配置文件，默认路径 `../conf/config.toml`。业务如需更换配置文件路径，需在调用New前显式赋值`going/config`包级变量`ConfPath`
- 配置struct为`going/client/req`包的[Conf struct](https://git.woa.com/tme/going/blob/master/client/req/request.go)
- 配置样例:
  ```yaml
  ["tme.microservices.test.1"]
  #address="ip://9.157.66.157:41018"
  timeout="2000ms"
  
  [CommonCount]
  address = "gl5://553153:65536"
  network = "tcp"
  timeout= "800ms"
  reqType=2
  
  [Score]
  address = "ip://9.157.66.157:41018"
  network = "tcp"
  timeout= "800ms"
  reqType=2
  ```

## 4、客户端调用流程

1. `New`创建客户端代理对象
2. 根据`CallDesc`参数从配置文件加载相应配置
3. 序列化**第三方协议首部**+**应用body**
4. 调用`going/client`触发网络请求
   - TCP：长连接类型复用连接池
   - UDP
   - Unix Socket
5. 解析响应
6. 反序列化**第三方协议首部**+**应用body**

最近更新于 2021.06.17
