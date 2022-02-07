package listen

import (
	"context"
	"looklook/app/order/cmd/mq/internal/config"
	"looklook/app/order/cmd/mq/internal/svc"

	"github.com/zeromicro/go-zero/core/service"
)

//返回所有消费者
func Mqs(c config.Config) []service.Service {

	svcContext := svc.NewServiceContext(c)
	ctx := context.Background()

	var services []service.Service

	//kq ：消息队列.
	services = append(services, KqMqs(c, ctx, svcContext)...)
	//asynq ： 延迟队列、定时任务
	services = append(services, AsynqMqs(c, ctx, svcContext)...)
	//other mq ....

	return services
}
