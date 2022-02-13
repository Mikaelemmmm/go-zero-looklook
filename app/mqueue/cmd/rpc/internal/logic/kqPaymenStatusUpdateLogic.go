package logic

import (
	"context"

	"looklook/app/mqueue/cmd/rpc/internal/svc"
	"looklook/app/mqueue/cmd/rpc/pb"
	"looklook/common/kqueue"

	"github.com/zeromicro/go-zero/core/logx"
)

type KqPaymenStatusUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewKqPaymenStatusUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KqPaymenStatusUpdateLogic {
	return &KqPaymenStatusUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 支付流水状态变更发送到kq..
func (l *KqPaymenStatusUpdateLogic) KqPaymenStatusUpdate(in *pb.KqPaymenStatusUpdateReq) (*pb.KqPaymenStatusUpdateResp, error) {

	m := kqueue.ThirdPaymentUpdatePayStatusNotifyMessage{
		OrderSn:   in.OrderSn,
		PayStatus: in.PayStatus,
	}

	if err := l.svcCtx.KqueueClient.Push(kqueue.PAYMENT_UPDATE_PAYSTATUS, m); err != nil {
		return nil, err
	}

	return &pb.KqPaymenStatusUpdateResp{}, nil
}
