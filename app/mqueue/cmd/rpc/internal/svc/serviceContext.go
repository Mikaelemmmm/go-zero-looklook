package svc

import (
	"looklook/app/mqueue/cmd/rpc/internal/config"
	"looklook/common/kqueue"

	"github.com/hibiken/asynq"
)

type ServiceContext struct {
	Config config.Config

	KqueueClient kqueue.KqueueClient
	AsynqClient  *asynq.Client
}

func NewServiceContext(c config.Config) *ServiceContext {

	svc := &ServiceContext{
		Config:       c,
		KqueueClient: kqueue.NewKqueueSvcClient(c.KqServerConf.Brokers),
		AsynqClient:  asynq.NewClient(asynq.RedisClientOpt{Addr: c.Redis.Host, Password: c.Redis.Pass}),
	}

	return svc
}
