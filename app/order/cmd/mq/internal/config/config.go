package config

import (
	"github.com/tal-tech/go-queue/kq"
	"github.com/tal-tech/go-zero/core/service"
	"github.com/tal-tech/go-zero/core/stores/redis"
	"github.com/tal-tech/go-zero/zrpc"
)

type Config struct {
	service.ServiceConf

	Redis redis.RedisConf

	//kq
	PaymentUpdateStatusConf kq.KqConf

	//rpc
	OrderRpcConf      zrpc.RpcClientConf
	MqueueRpcConf     zrpc.RpcClientConf
	UsercenterRpcConf zrpc.RpcClientConf
}
