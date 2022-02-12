package system

import (
	v1 "looklook/admin/api/v1"
	"github.com/gin-gonic/gin"
)

type JwtRouter struct{}

func (s *JwtRouter) InitJwtRouter(Router *gin.RouterGroup) {
	jwtRouter := Router.Group("jwt")
	jwtApi := v1.ApiGroupApp.SystemApiGroup.JwtApi
	{
		jwtRouter.POST("jsonInBlacklist", jwtApi.JsonInBlacklist) // jwt加入黑名单
	}
}
