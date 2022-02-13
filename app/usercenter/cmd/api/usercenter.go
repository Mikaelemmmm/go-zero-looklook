package main

import (
	"flag"
	"fmt"
	"looklook/app/usercenter/cmd/api/internal/config"
	"looklook/app/usercenter/cmd/api/internal/handler"
	"looklook/app/usercenter/cmd/api/internal/svc"
	"looklook/common/middleware"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/usercenter.yaml", "the config file")

func main() {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	ctx := svc.NewServiceContext(c)
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 全局中间件
	// 将nginx网关验证后的userId设置到ctx中
	server.Use(middleware.NewSetUidToCtxMiddleware().Handle)

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
