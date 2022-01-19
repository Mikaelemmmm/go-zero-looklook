package main

import (
	"flag"
	"github.com/tal-tech/go-zero/core/prometheus"

	"looklook/app/order/cmd/mq/internal/config"
	"looklook/app/order/cmd/mq/internal/listen"

	"github.com/tal-tech/go-zero/core/conf"
	"github.com/tal-tech/go-zero/core/service"
)

var configFile = flag.String("f", "etc/order.yaml", "Specify the config file")

func main() {
	flag.Parse()
	var c config.Config

	conf.MustLoad(*configFile, &c)
	prometheus.StartAgent(c.Prometheus)

	serviceGroup := service.NewServiceGroup()
	defer serviceGroup.Stop()

	for _, mq := range listen.Mqs(c) {
		serviceGroup.Add(mq)
	}

	serviceGroup.Start()

}
