package svc

import (
	"fmt"
	"github.com/hibiken/asynq"
	"looklook/app/mqueue/cmd/job/internal/config"
)

type AsynqServer struct {
	Config config.Config
}

func newAsynqServer(c config.Config) *asynq.Server {
	// 设置起初线程数为：1
	concurrency := 1
	if c.AsynqConf.MqConcurrency != 0 {
		concurrency = int(c.AsynqConf.MqConcurrency)
	}
	// 指定一个queue的名称（可写入common中）
	queueName := "queueTest"
	return asynq.NewServer(

		asynq.RedisClientOpt{Addr: c.Redis.Host, Password: c.Redis.Pass},

		asynq.Config{
			IsFailure: func(err error) bool {
				fmt.Printf("asynq server exec task IsFailure ======== >>>>>>>>>>>  err : %+v \n", err)
				return true
			},
			// 将线程控制在k8s configmap中管理，默认为1
			Concurrency: concurrency, //max concurrent process job task num
			// Optionally specify multiple queues with different priority.
			// See the godoc for other configuration options
			// ↑ 以上为官方说明
			// 建议使用队列名，如果不指定queue名称，则会自动进入到queue名为default的队列中，只是按type的名称区别
			// 多业务场景下，不推荐不指定queue的Name
			Queues: map[string]int{queueName: 1}, // 这里有一个权重如果给1则100%
		},
	)
}
