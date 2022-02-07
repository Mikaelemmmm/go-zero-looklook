package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	WxMiniConf WxMiniConf
	WxPayConf  WxPayConf

	PaymentRpcConf    zrpc.RpcClientConf
	OrderRpcConf      zrpc.RpcClientConf
	UsercenterRpcConf zrpc.RpcClientConf
}
