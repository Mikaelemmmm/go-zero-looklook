package svc

import (
	"github.com/hibiken/asynq"
	"github.com/silenceper/wechat/v2/miniprogram"
	"github.com/zeromicro/go-zero/zrpc"
	"looklook/app/mqueue/cmd/job/internal/config"
	"looklook/app/order/cmd/rpc/order"
	"looklook/app/usercenter/cmd/rpc/usercenter"
)

type ServiceContext struct {
	Config config.Config
	AsynqServer *asynq.Server
	MiniProgram *miniprogram.MiniProgram

	OrderRpc order.Order
	UsercenterRpc usercenter.Usercenter
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		AsynqServer:newAsynqServer(c),
		MiniProgram:newMiniprogramClient(c),
		OrderRpc:order.NewOrder(zrpc.MustNewClient(c.OrderRpcConf)),
		UsercenterRpc:usercenter.NewUsercenter(zrpc.MustNewClient(c.UsercenterRpcConf)),
	}
}



