package config

import (
	"github.com/tal-tech/go-queue/kq"
	"github.com/tal-tech/go-zero/core/service"
)

type Config struct {
	service.ServiceConf
	//kq
	SendWxMiniTplMessageConf kq.KqConf
	WxMiniConf               WxMiniConf
}
