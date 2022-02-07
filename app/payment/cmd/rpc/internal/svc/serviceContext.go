package svc

import (
	"looklook/app/mqueue/cmd/rpc/mqueue"
	"looklook/app/payment/cmd/rpc/internal/config"
	"looklook/app/payment/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config            config.Config
	ThirdPaymentModel model.ThirdPaymentModel
	MqueueRpc         mqueue.Mqueue
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,

		ThirdPaymentModel: model.NewThirdPaymentModel(sqlx.NewMysql(c.DB.DataSource), c.Cache),

		MqueueRpc: mqueue.NewMqueue(zrpc.MustNewClient(c.MqueueRpcConf)),
	}
}
