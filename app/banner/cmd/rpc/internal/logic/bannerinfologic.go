package logic

import (
	"context"

	"looklook/app/banner/cmd/rpc/internal/svc"
	"looklook/app/banner/cmd/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type BannerInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBannerInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BannerInfoLogic {
	return &BannerInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 广告详情
func (l *BannerInfoLogic) BannerInfo(in *pb.BannerInfoReq) (*pb.BannerInfoResp, error) {


	return &pb.BannerInfoResp{
		Banner: &pb.Banner{
			Id: 1,
			Title: "与admin（gin-view-admin）模块测试使用数据",
			Forward: "https://github.com/Mikaelemmmm/go-zero-looklook",
			Img: "https://github.com/Mikaelemmmm/go-zero-looklook",
		},
	}, nil
}
