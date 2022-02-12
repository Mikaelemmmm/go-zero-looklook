package banner

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	"looklook/admin/model/banner/response"
	"looklook/admin/rpc"
	"looklook/app/banner/cmd/rpc/pb"
	"looklook/common/xerr"
)

type BannerService struct {
}


// GetBanner 根据id获取Banner记录
// Author [piexlmax](https://github.com/piexlmax)
func (bannerService *BannerService)GetBanner(id int64) (err error, banner *response.GetBanner) {

	//调用go-zero的 rpc请求
	bannerInfoResp , err:= rpc.GetClient().BannerRpc.BannerInfo(context.Background(),&pb.BannerInfoReq{
		Id: 1,
	})
	if err != nil{
		return xerr.NewErrMsg("查询信息失败"),nil
	}

	var resp response.GetBanner
	_ = copier.Copy(&resp,bannerInfoResp.Banner)

	fmt.Printf("resp 222: %+v \n",resp)
	return nil,&resp
}
