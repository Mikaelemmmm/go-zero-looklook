package config

import (
	"github.com/tal-tech/go-zero/rest"
	"github.com/tal-tech/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	OrderRpcConf   zrpc.RpcClientConf
	PaymentRpcConf zrpc.RpcClientConf
}
