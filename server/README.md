参考 net、gin 的设计

todo:
1. 当 server handler 中发生了 panic 之后，整个服务都会 down 掉，需要加 recover 处理
