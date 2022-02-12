package banner

import (
	"github.com/gin-gonic/gin"
	"looklook/admin/model/banner/request"
	"looklook/admin/model/common/response"
	"looklook/admin/service"
)


type BannerApi struct {
}

var bannerService = service.ServiceGroupApp.BannerServiceGroup.BannerService



// FindBanner 用id查询Banner
// @Tags Banner
// @Summary 用id查询Banner
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query autocode.Banner true "用id查询Banner"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /banner/findBanner [get]
func (bannerApi *BannerApi) FindBanner(c *gin.Context) {

	var banner request.GetBanner
	_ = c.ShouldBindJSON(&banner)

	if err ,resp := bannerService.GetBanner(banner.Id);err!= nil {
		response.FailWithMessage("创建失败", c)
	} else {
		response.OkWithData(resp, c)
	}
}
