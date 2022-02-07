package svc

import (
	"looklook/app/identity/cmd/rpc/identity"
	"looklook/app/usercenter/cmd/rpc/internal/config"
	"looklook/app/usercenter/model"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	//本地服务
	Config      config.Config
	RedisClient *redis.Redis

	//rpc服务
	IdentityRpc identity.Identity

	//model
	UserModel     model.UserModel
	UserAuthModel model.UserAuthModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:      c,
		RedisClient: redis.NewRedis(c.Redis.Host, c.Redis.Type, c.Redis.Pass),

		IdentityRpc: identity.NewIdentity(zrpc.MustNewClient(c.IdentityRpcConf)),

		UserAuthModel: model.NewUserAuthModel(sqlx.NewMysql(c.DB.DataSource), c.Cache),
		UserModel:     model.NewUserModel(sqlx.NewMysql(c.DB.DataSource), c.Cache),
	}
}
