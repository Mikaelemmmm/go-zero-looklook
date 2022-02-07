package main

import (
	"flag"
	"fmt"

	"looklook/app/mqueue/cmd/rpc/internal/config"
	"looklook/app/mqueue/cmd/rpc/internal/server"
	"looklook/app/mqueue/cmd/rpc/internal/svc"
	"looklook/app/mqueue/cmd/rpc/pb"
	"looklook/common/interceptor/rpcserver"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/mqueue.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)
	srv := server.NewMqueueServer(ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterMqueueServer(grpcServer, srv)

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	//rpc log.

	s.AddUnaryInterceptors(rpcserver.LoggerInterceptor)

	defer func() {
		s.Stop()

		//关闭asynq客户端
		ctx.AsynqClient.Close()
	}()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
