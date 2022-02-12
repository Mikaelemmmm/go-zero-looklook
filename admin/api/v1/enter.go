package v1

import (
	"looklook/admin/api/v1/autocode"
	"looklook/admin/api/v1/banner"
	"looklook/admin/api/v1/example"
	"looklook/admin/api/v1/system"
)

type ApiGroup struct {
	SystemApiGroup   system.ApiGroup
	ExampleApiGroup  example.ApiGroup
	AutoCodeApiGroup autocode.ApiGroup
	BannerApiGroup  banner.ApiGroup
}

var ApiGroupApp = new(ApiGroup)
