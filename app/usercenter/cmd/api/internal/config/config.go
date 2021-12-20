package config

import (
	"github.com/tal-tech/go-zero/rest"
	"github.com/tal-tech/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	JwtAuth struct {
		AccessSecret string
	}
	//
	WxMiniConf        WxMiniConf
	UsercenterRpcConf zrpc.RpcClientConf
	IdentityRpcConf   zrpc.RpcClientConf
}
