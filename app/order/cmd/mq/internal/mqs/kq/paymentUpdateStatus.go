package kq

import (
	"context"
	"encoding/json"
	"fmt"

	"looklook/app/mqueue/cmd/rpc/mqueue"
	"looklook/app/order/cmd/mq/internal/svc"
	"looklook/app/order/cmd/rpc/order"
	"looklook/app/order/model"
	paymentModel "looklook/app/payment/model"
	"looklook/app/usercenter/cmd/rpc/usercenter"
	usercenterModel "looklook/app/usercenter/model"
	"looklook/common/kqueue"
	"looklook/common/tool"
	"looklook/common/wxminisub"
	"looklook/common/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

/**
监听支付流水状态变更通知消息队列
*/
type PaymentUpdateStatusMq struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPaymentUpdateStatusMq(ctx context.Context, svcCtx *svc.ServiceContext) *PaymentUpdateStatusMq {
	return &PaymentUpdateStatusMq{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PaymentUpdateStatusMq) Consume(_, val string) error {

	fmt.Printf(" PaymentUpdateStatusMq Consume val : %s \n", val)
	//解析数据
	var message kqueue.ThirdPaymentUpdatePayStatusNotifyMessage
	if err := json.Unmarshal([]byte(val), &message); err != nil {
		logx.WithContext(l.ctx).Error("PaymentUpdateStatusMq->Consume Unmarshal err : %v , val : %s", err, val)
		return err
	}

	//执行业务..
	if err := l.execService(message); err != nil {
		logx.WithContext(l.ctx).Error("PaymentUpdateStatusMq->execService  err : %v , val : %s , message:%+v", err, val, message)
		return err
	}

	return nil
}

//执行业务
func (l *PaymentUpdateStatusMq) execService(message kqueue.ThirdPaymentUpdatePayStatusNotifyMessage) error {

	orderTradeState := l.getOrderTradeStateByPaymentTradeState(message.PayStatus)
	if orderTradeState != -99 {
		//更新订单状态
		resp, err := l.svcCtx.OrderRpc.UpdateHomestayOrderTradeState(l.ctx, &order.UpdateHomestayOrderTradeStateReq{
			Sn:         message.OrderSn,
			TradeState: orderTradeState,
		})
		if err != nil {
			return errors.Wrapf(xerr.NewErrMsg("更新订单状态失败"), " err : %v ,message:%+v", err, message)
		}

		//发送短信、微信等通知用户
		l.notifyUser(resp.Sn, resp.TradeCode, resp.Title, resp.OrderTotalPrice, resp.LiveStartDate, resp.LiveEndDate, resp.UserId)

	}

	return nil
}

//根据支付状态获取订单状态.
func (l *PaymentUpdateStatusMq) getOrderTradeStateByPaymentTradeState(thirdPaymentPayStatus int64) int64 {

	switch thirdPaymentPayStatus {
	case paymentModel.ThirdPaymentPayTradeStateSuccess:
		return model.HomestayOrderTradeStateWaitUse
	case paymentModel.ThirdPaymentPayTradeStateRefund:
		return model.HomestayOrderTradeStateRefund
	default:
		return -99
	}

}

//发送小程序模版消息通知用户
func (l *PaymentUpdateStatusMq) notifyUser(sn, code, title string, orderTotalPrice, liveStartDate, liveEndDate, userId int64) {

	usercenterResp, err := l.svcCtx.UsercenterRpc.GetUserAuthByUserId(l.ctx, &usercenter.GetUserAuthByUserIdReq{
		UserId:   userId,
		AuthType: usercenterModel.UserAuthTypeSmallWX,
	})
	if err != nil {
		logx.WithContext(l.ctx).Errorf("PaymentUpdateStatusMq -> notifyUser err:%v , sn :%s , code : %s , userId : %d ", err, sn, code, userId)
	}
	if usercenterResp.UserAuth == nil || len(usercenterResp.UserAuth.AuthKey) == 0 {
		logx.WithContext(l.ctx).Errorf("PaymentUpdateStatusMq -> notifyUser 未保存用户openid信息， 数据异常  sn :%s , code : %s , userId : %d ", sn, code, userId)
	}

	openId := usercenterResp.UserAuth.AuthKey

	//发送小程序订单支付成功订阅消息..
	orderPaySuccessDataParam := wxminisub.OrderPaySuccessDataParam{
		Sn:            sn,
		GoodsName:     title,
		PayTotal:      fmt.Sprintf("%.2f", tool.Fen2Yuan(orderTotalPrice)),
		LiveStateDate: "2021-12-07",
		LiveEndDate:   "2021-12-08",
	}
	if _, err = l.svcCtx.MqueueRpc.SendWxMiniSubMessage(l.ctx, &mqueue.SendWxMiniSubMessageReq{
		TemplateID: wxminisub.OrderPaySuccessTemplateID,
		Openid:     openId,
		Data:       wxminisub.OrderPaySuccessData(orderPaySuccessDataParam),
	}); err != nil {
		logx.WithContext(l.ctx).Errorf("发送小程序订单支付成功订阅消息失败 orderPaySuccessDataParam : %+v \n", orderPaySuccessDataParam)
	}

	//发送小程序入驻须知订阅消息
	orderPaySuccessLiveKnowDataParam := wxminisub.OrderPaySuccessLiveKnowDataParam{
		LiveStartDate: "2021-12-07",
		LiveEndDate:   "2021-12-08",
		TradeCode:     code,
		Remark:        "请到商家出示【入住密码】进行入住",
	}
	_, err = l.svcCtx.MqueueRpc.SendWxMiniSubMessage(l.ctx, &mqueue.SendWxMiniSubMessageReq{
		TemplateID: wxminisub.OrderPaySuccessLiveKnowTemplateID,
		Openid:     openId,
		Data:       wxminisub.OrderPaySuccessLiveKnowData(orderPaySuccessLiveKnowDataParam),
	})
	if err != nil {
		logx.WithContext(l.ctx).Errorf("发送小程序入驻须知订阅消息失败 orderPaySuccessLiveKnowDataParam : %+v \n", orderPaySuccessLiveKnowDataParam)
	}
}
