package svc

import (
	"looklook/app/mqueue/cmd/rpc/mqueue"
	"looklook/app/order/cmd/mq/internal/config"
	"looklook/app/order/cmd/rpc/order"
	"looklook/app/usercenter/cmd/rpc/usercenter"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	OrderRpc      order.Order
	MqueueRpc     mqueue.Mqueue
	UsercenterRpc usercenter.Usercenter
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,

		OrderRpc:      order.NewOrder(zrpc.MustNewClient(c.OrderRpcConf)),
		MqueueRpc:     mqueue.NewMqueue(zrpc.MustNewClient(c.MqueueRpcConf)),
		UsercenterRpc: usercenter.NewUsercenter(zrpc.MustNewClient(c.UsercenterRpcConf)),
	}
}
