package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	JwtAuth struct {
		AccessSecret string
		AccessExpire int64
	}
}
