package svc

import (
	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-queue/kq"
	"looklook/app/mqueue/cmd/rpc/internal/config"
)

type ServiceContext struct {
	Config config.Config

	AsynqClient  *asynq.Client

	KqueuePaymentUpdatePayStatusClient *kq.Pusher
	KqueueSendWxMiniTplMessageClient   *kq.Pusher

}

func NewServiceContext(c config.Config) *ServiceContext {

	svc := &ServiceContext{
		Config:       c,
		AsynqClient:  asynq.NewClient(asynq.RedisClientOpt{Addr: c.Redis.Host, Password: c.Redis.Pass}),
		KqueuePaymentUpdatePayStatusClient: kq.NewPusher(c.KqPaymentUpdatePayStatusConf.Brokers,c.KqPaymentUpdatePayStatusConf.Topic),
		KqueueSendWxMiniTplMessageClient: kq.NewPusher(c.KqSendWxMiniTplMessageConf.Brokers,c.KqSendWxMiniTplMessageConf.Topic),
	}

	return svc
}
