package listen

import (
	"context"
	"looklook/app/message/cmd/mq/internal/config"
	"looklook/app/message/cmd/mq/internal/svc"

	"github.com/zeromicro/go-zero/core/service"
)

//asynq
func AsynqMqs(c config.Config, ctx context.Context, svcContext *svc.ServiceContext) []service.Service {

	return []service.Service{
		//监听消费流水状态变更
		//.....
	}

}
