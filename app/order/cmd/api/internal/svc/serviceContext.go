package svc

import (
	"looklook/app/order/cmd/api/internal/config"
	"looklook/app/order/cmd/rpc/order"
	"looklook/app/payment/cmd/rpc/payment"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	OrderRpc   order.Order
	PaymentrPC payment.Payment
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,

		OrderRpc:   order.NewOrder(zrpc.MustNewClient(c.OrderRpcConf)),
		PaymentrPC: payment.NewPayment(zrpc.MustNewClient(c.PaymentRpcConf)),
	}
}
