package svc

import (
	"github.com/hibiken/asynq"
)

//创建asynq client.
func newAsynqClient(redisHost string, redisPassword string) *asynq.Client {

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisHost, Password: redisPassword})

	return client
}
