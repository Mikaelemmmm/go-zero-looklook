package logic

import (
	"context"

	"looklook/app/travel/cmd/rpc/internal/svc"
	"looklook/app/travel/cmd/rpc/pb"
	"looklook/app/travel/model"
	"looklook/common/xerr"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type HomestayDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHomestayDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HomestayDetailLogic {
	return &HomestayDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 民宿详情.
func (l *HomestayDetailLogic) HomestayDetail(in *pb.HomestayDetailReq) (*pb.HomestayDetailResp, error) {

	homestay, err := l.svcCtx.HomestayModel.FindOne(in.Id)
	if err != nil && err != model.ErrNotFound {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), " id : %d ", in.Id)
	}

	var pbHomestay pb.Homestay
	if homestay != nil {
		//整合民宿详情
		_ = copier.Copy(&pbHomestay, homestay)
	}

	return &pb.HomestayDetailResp{
		Homestay: &pbHomestay,
	}, nil

}
