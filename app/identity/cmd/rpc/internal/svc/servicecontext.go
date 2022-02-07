package svc

import (
	"looklook/app/identity/cmd/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ServiceContext struct {
	Config      config.Config
	RedisClient *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:      c,
		RedisClient: redis.NewRedis(c.Redis.Host, c.Redis.Type, c.Redis.Pass),
	}
}
