package service

import (
	"looklook/admin/service/autocode"
	"looklook/admin/service/banner"
	"looklook/admin/service/example"
	"looklook/admin/service/system"
)

type ServiceGroup struct {
	SystemServiceGroup   system.ServiceGroup
	ExampleServiceGroup  example.ServiceGroup
	AutoCodeServiceGroup autocode.ServiceGroup
	BannerServiceGroup   banner.ServiceGroup
}

var ServiceGroupApp = new(ServiceGroup)
