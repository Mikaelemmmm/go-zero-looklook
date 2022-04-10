package svc

import (
	"github.com/hibiken/asynq"
	"looklook/app/order/cmd/rpc/internal/config"
)

//create asynq client.
func newAsynqClient(c config.Config) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{Addr: c.Redis.Host, Password: c.Redis.Pass})
}
