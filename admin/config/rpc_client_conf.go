package config


import (
	"github.com/zeromicro/go-zero/zrpc"
)
//rpcConf
type RpcConf struct{
	BannerRpcConf zrpc.RpcClientConf `mapstructure:"bannerRpcConf" json:"bannerRpcConf" yaml:"bannerRpcConf"`
}
