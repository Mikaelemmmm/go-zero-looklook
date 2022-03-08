package logic

import (
	"context"
	"github.com/pkg/errors"
	"looklook/common/xerr"
	"encoding/json"

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

	//2、序列化
	body, err := json.Marshal(m)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrMsg("kq kqPaymenStatusUpdateLogic  task marshal error "), "kq kqPaymenStatusUpdateLogic  task marshal error , v : %+v", m)
	}

	if err := l.svcCtx.KqueuePaymentUpdatePayStatusClient.Push(string(body)); err != nil {
		return nil, err
	}

	return &pb.KqPaymenStatusUpdateResp{}, nil
}
