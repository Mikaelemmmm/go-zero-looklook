package config

import (
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/service"
)

type Config struct {
	service.ServiceConf
	//kq
	SendWxMiniTplMessageConf kq.KqConf
	WxMiniConf               WxMiniConf
}
