package logic

import (
	"context"
	"time"

	"looklook/app/mqueue/cmd/rpc/internal/svc"
	"looklook/app/mqueue/cmd/rpc/pb"
	"looklook/common/asynqmq"
	"looklook/common/xerr"

	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type AqDeferHomestayOrderCloseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAqDeferHomestayOrderCloseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AqDeferHomestayOrderCloseLogic {
	return &AqDeferHomestayOrderCloseLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 添加民宿订单延迟关闭到asynq队列
func (l *AqDeferHomestayOrderCloseLogic) AqDeferHomestayOrderClose(in *pb.AqDeferHomestayOrderCloseReq) (*pb.AqDeferHomestayOrderCloseResp, error) {

	task, err := asynqmq.NewHomestayOrderCloseTask(in.Sn)
	if err != nil {
		return nil, err
	}

	_, err = l.svcCtx.AsynqClient.Enqueue(task, asynq.ProcessIn(20*time.Minute))
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrMsg("添加民宿订单到延迟队列失败"), "添加民宿订单到延迟队列失败 sn:%s ,err:%v", in.Sn, err)
	}

	return &pb.AqDeferHomestayOrderCloseResp{}, nil
}
