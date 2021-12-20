package deferMq

import (
	"context"
	"encoding/json"
	"looklook/app/order/cmd/rpc/order"
	"looklook/app/order/model"
	"looklook/common/asynqmq"
	"looklook/common/xerr"

	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
)

func (l *AsynqTask) closeHomestayOrderStateMqHandler(ctx context.Context, t *asynq.Task) error {

	var p asynqmq.HomestayOrderCloseTaskPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return errors.Wrapf(xerr.NewErrMsg("解析asynq task payload err"), "closeHomestayOrderStateMqHandler payload err:%v, payLoad:%+v", err, t.Payload())
	}

	resp, err := l.svcCtx.OrderRpc.HomestayOrderDetail(ctx, &order.HomestayOrderDetailReq{
		Sn: p.Sn,
	})
	if err != nil || resp.HomestayOrder == nil {
		return errors.Wrapf(xerr.NewErrMsg("获取订单失败"), "closeHomestayOrderStateMqHandler 获取订单失败 or 订单不存在 err:%v, sn:%s ,HomestayOrder : %+v", err, p.Sn, resp.HomestayOrder)
	}

	if resp.HomestayOrder.TradeState == model.HomestayOrderTradeStateWaitPay {
		_, err := l.svcCtx.OrderRpc.UpdateHomestayOrderTradeState(ctx, &order.UpdateHomestayOrderTradeStateReq{
			Sn:         p.Sn,
			TradeState: model.HomestayOrderTradeStateCancel,
		})
		if err != nil {
			return errors.Wrapf(xerr.NewErrMsg("关闭订单失败"), "closeHomestayOrderStateMqHandler 关闭订单失败  err:%v, sn:%s ", err, p.Sn)
		}
	}

	return nil
}
