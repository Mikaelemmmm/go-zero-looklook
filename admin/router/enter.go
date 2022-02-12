package router

import (
	"looklook/admin/router/autocode"
	"looklook/admin/router/banner"
	"looklook/admin/router/example"
	"looklook/admin/router/system"
)

type RouterGroup struct {
	System   system.RouterGroup
	Example  example.RouterGroup
	Autocode autocode.RouterGroup
	Banner  banner.RouterGroup
}

var RouterGroupApp = new(RouterGroup)
