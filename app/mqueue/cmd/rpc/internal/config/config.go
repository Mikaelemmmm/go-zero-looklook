package config

import "github.com/tal-tech/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	KqServerConf struct {
		Brokers []string
	}
}
