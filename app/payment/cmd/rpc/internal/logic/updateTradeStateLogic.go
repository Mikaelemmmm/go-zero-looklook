package logic

import (
	"context"
	"time"

	"looklook/app/mqueue/cmd/rpc/mqueue"
	"looklook/app/payment/cmd/rpc/internal/svc"
	"looklook/app/payment/cmd/rpc/pb"
	"looklook/app/payment/model"
	"looklook/common/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTradeStateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTradeStateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTradeStateLogic {
	return &UpdateTradeStateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新交易状态
func (l *UpdateTradeStateLogic) UpdateTradeState(in *pb.UpdateTradeStateReq) (*pb.UpdateTradeStateResp, error) {

	//1、流水记录确认.
	thirdPayment, err := l.svcCtx.ThirdPaymentModel.FindOneBySn(in.Sn)
	if err != nil && err != model.ErrNotFound {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "更新交易状态 ，根据流水单号查询流水db异常 sn : %s", in.Sn)
	}

	if thirdPayment == nil {
		return nil, errors.Wrapf(xerr.NewErrMsg("该流水记录不存在"), " sn : %s", in.Sn)
	}

	//2、判断状态
	if in.PayStatus == model.ThirdPaymentPayTradeStateSuccess || in.PayStatus == model.ThirdPaymentPayTradeStateFAIL {
		//想要修改为支付成功、失败场景

		if thirdPayment.PayStatus != model.ThirdPaymentPayTradeStateWait {
			return &pb.UpdateTradeStateResp{}, nil
		}

	} else if in.PayStatus == model.ThirdPaymentPayTradeStateRefund {
		//想要修改为退款成功场景

		if thirdPayment.PayStatus != model.ThirdPaymentPayTradeStateSuccess {
			return nil, errors.Wrapf(xerr.NewErrMsg("只有支付成功的订单才能退款"), "修改支付流水记录为退款失败，当前支付流水未支付成功无法退款 in : %+v", in)
		}
	} else {
		return nil, errors.Wrapf(xerr.NewErrMsg("当前不支持此状态"), "修改支付流水状态不支持  in : %+v", in)
	}

	//3、更新.
	thirdPayment.TradeState = in.TradeState
	thirdPayment.TransactionId = in.TransactionId
	thirdPayment.TradeType = in.TradeType
	thirdPayment.TradeStateDesc = in.TradeStateDesc
	thirdPayment.PayStatus = in.PayStatus
	thirdPayment.PayTime = time.Unix(in.PayTime, 0)
	if err := l.svcCtx.ThirdPaymentModel.Update(nil, thirdPayment); err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), " 更新流水状态失败 err:%v ", err)
	}

	//4、通知其他服务
	_, _ = l.svcCtx.MqueueRpc.KqPaymenStatusUpdate(l.ctx, &mqueue.KqPaymenStatusUpdateReq{
		OrderSn:   thirdPayment.OrderSn,
		PayStatus: in.PayStatus,
	})

	return &pb.UpdateTradeStateResp{}, nil
}
