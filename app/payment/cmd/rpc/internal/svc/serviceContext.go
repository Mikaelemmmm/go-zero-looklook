package svc

import (
	"github.com/zeromicro/go-queue/kq"
	"looklook/app/payment/cmd/rpc/internal/config"
	"looklook/app/payment/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config            config.Config
	ThirdPaymentModel model.ThirdPaymentModel
	KqueuePaymentUpdatePayStatusClient *kq.Pusher
}

func NewServiceContext(c config.Config) *ServiceContext {

	sqlConn := sqlx.NewMysql(c.DB.DataSource)

	return &ServiceContext{
		Config: c,
		ThirdPaymentModel: model.NewThirdPaymentModel(sqlConn, c.Cache),
		KqueuePaymentUpdatePayStatusClient: kq.NewPusher(c.KqPaymentUpdatePayStatusConf.Brokers,c.KqPaymentUpdatePayStatusConf.Topic),
	}
}
