package listen

import (
	"context"
	"looklook/app/order/cmd/mq/internal/config"
	kqMq "looklook/app/order/cmd/mq/internal/mqs/kq"
	"looklook/app/order/cmd/mq/internal/svc"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/service"
)

//pub sub use kq (kafka)
func KqMqs(c config.Config, ctx context.Context, svcContext *svc.ServiceContext) []service.Service {

	return []service.Service{
		//Listening for changes in consumption flow status
		kq.MustNewQueue(c.PaymentUpdateStatusConf, kqMq.NewPaymentUpdateStatusMq(ctx, svcContext)),
		//.....
	}

}
