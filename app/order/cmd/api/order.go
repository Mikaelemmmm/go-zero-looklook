package main

import (
	"flag"

	"looklook/app/order/cmd/api/internal/config"
	"looklook/app/order/cmd/api/internal/handler"
	"looklook/app/order/cmd/api/internal/svc"
	"looklook/common/middleware"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/order.yaml", "the config file")

func main() {

	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 全局中间件
	//将nginx网关验证后的userId设置到ctx中.
	server.Use(middleware.NewSetUidToCtxMiddleware().Handle)

	handler.RegisterHandlers(server, ctx)

	server.Start()
}
