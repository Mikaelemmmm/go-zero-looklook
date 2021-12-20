package svc

import (
	"looklook/app/identity/cmd/api/internal/config"
	"looklook/app/identity/cmd/rpc/identity"

	"github.com/tal-tech/go-zero/core/stores/redis"
	"github.com/tal-tech/go-zero/zrpc"
)

type ServiceContext struct {
	Config      config.Config
	RedisClient *redis.Redis
	IdentityRpc identity.Identity
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:      c,
		IdentityRpc: identity.NewIdentity(zrpc.MustNewClient(c.IdentityRpcConf)),
	}
}
