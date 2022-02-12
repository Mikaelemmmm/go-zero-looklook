package banner

import (
	"github.com/gin-gonic/gin"
	"looklook/admin/api/v1"
	"looklook/admin/middleware"
)

type BannerRouter struct {
}

// InitBannerRouter 初始化 Banner 路由信息
func (s *BannerRouter) InitBannerRouter(Router *gin.RouterGroup) {
	bannerRouter := Router.Group("banner").Use(middleware.OperationRecord())
	var bannerApi = v1.ApiGroupApp.BannerApiGroup.BannerApi
	{
		bannerRouter.POST("find", bannerApi.FindBanner)   // 查询banner
	}
}
