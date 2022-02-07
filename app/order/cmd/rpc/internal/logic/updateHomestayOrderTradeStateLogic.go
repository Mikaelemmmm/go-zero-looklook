package logic

import (
	"context"

	"looklook/app/order/cmd/rpc/internal/svc"
	"looklook/app/order/cmd/rpc/pb"
	"looklook/app/order/model"
	"looklook/common/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateHomestayOrderTradeStateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateHomestayOrderTradeStateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateHomestayOrderTradeStateLogic {
	return &UpdateHomestayOrderTradeStateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新民宿订单状态
func (l *UpdateHomestayOrderTradeStateLogic) UpdateHomestayOrderTradeState(in *pb.UpdateHomestayOrderTradeStateReq) (*pb.UpdateHomestayOrderTradeStateResp, error) {

	// 1、查询当前订单
	homestayOrder, err := l.svcCtx.HomestayOrderModel.FindOneBySn(in.Sn)
	if err != nil && err != model.ErrNotFound {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "err : %v , in:%+v", err, in)
	}
	if homestayOrder == nil {
		return nil, errors.Wrapf(xerr.NewErrMsg("订单不存在"), "in : %+v", in)
	}

	if homestayOrder.TradeState == in.TradeState {
		return &pb.UpdateHomestayOrderTradeStateResp{}, nil
	}

	// 2、校验订单状态
	if err := l.verifyOrderTradeState(in.TradeState, homestayOrder.TradeState); err != nil {
		return nil, errors.WithMessagef(err, " , in : %+v", in)
	}

	// 3、更新前状态判断.
	homestayOrder.TradeState = in.TradeState
	if err := l.svcCtx.HomestayOrderModel.Update(nil, homestayOrder); err != nil {
		return nil, errors.Wrapf(xerr.NewErrMsg("更新民宿订单状态失败"), "更新民宿订单状态失败 err:%v , in : %v", err, in)
	}

	return &pb.UpdateHomestayOrderTradeStateResp{
		Id:              homestayOrder.Id,
		UserId:          homestayOrder.UserId,
		Sn:              homestayOrder.Sn,
		TradeCode:       homestayOrder.TradeCode,
		Title:           homestayOrder.Title,
		LiveStartDate:   homestayOrder.LiveStartDate.Unix(),
		LiveEndDate:     homestayOrder.LiveEndDate.Unix(),
		OrderTotalPrice: homestayOrder.OrderTotalPrice,
	}, nil
}

// 更新民宿订单状态
func (l *UpdateHomestayOrderTradeStateLogic) verifyOrderTradeState(newTradeState, oldTradeState int64) error {
	if newTradeState == model.HomestayOrderTradeStateWaitPay {
		return errors.Wrapf(xerr.NewErrMsg("不支持更改此状态"),
			"不支持更改为待支付状态 newTradeState: %d, oldTradeState: %d",
			newTradeState,
			oldTradeState)
	}

	if newTradeState == model.HomestayOrderTradeStateCancel {

		if oldTradeState != model.HomestayOrderTradeStateWaitPay {
			return errors.Wrapf(xerr.NewErrMsg("只有待支付的订单才能被取消"),
				"只有待支付的订单才能被取消 newTradeState: %d, oldTradeState: %d",
				newTradeState,
				oldTradeState)
		}

	} else if newTradeState == model.HomestayOrderTradeStateWaitUse {
		if oldTradeState != model.HomestayOrderTradeStateWaitPay {
			return errors.Wrapf(xerr.NewErrMsg("只有待支付的订单才能更改为此状态"),
				"只有待支付的订单才能更改为未使用状态 newTradeState: %d, oldTradeState: %d",
				newTradeState,
				oldTradeState)
		}
	} else if newTradeState == model.HomestayOrderTradeStateUsed {
		if oldTradeState != model.HomestayOrderTradeStateWaitUse {
			return errors.Wrapf(xerr.NewErrMsg("只有未使用的订单才能更改为此状态"),
				"只有未使用的订单才能更改为已使用状态 newTradeState: %d, oldTradeState: %d",
				newTradeState,
				oldTradeState)
		}
	} else if newTradeState == model.HomestayOrderTradeStateRefund {
		if oldTradeState != model.HomestayOrderTradeStateWaitUse {
			return errors.Wrapf(xerr.NewErrMsg("只有未使用的订单才能更改为此状态"),
				"只有未使用的订单才能更改为退款状态 newTradeState: %d, oldTradeState: %d",
				newTradeState,
				oldTradeState)
		}
	} else if newTradeState == model.HomestayOrderTradeStateExpire {
		if oldTradeState != model.HomestayOrderTradeStateWaitUse {
			return errors.Wrapf(xerr.NewErrMsg("只有未使用的订单才能更改为此状态"),
				"只有未使用的订单才能更改为已过期状态 newTradeState: %d, oldTradeState: %d",
				newTradeState,
				oldTradeState)
		}
	}

	return nil
}
