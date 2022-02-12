import service from '@/utils/request'

// @Tags api
// @Summary 分页获取角色列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body modelInterface.PageInfo true "分页获取用户列表"
// @Success 200 {string} json "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /api/getApiList [post]
// {
//  page     int
//	pageSize int
// }
export const getBannerDetail = (data) => {
  return service({
    //这里演示，先写死
    url: '/banner/find',
    method: 'post',
    data
  })
}

