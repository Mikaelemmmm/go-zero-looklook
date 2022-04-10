package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-module/carbon/v2"
	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
	"github.com/silenceper/wechat/v2/miniprogram/subscribe"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"looklook/app/mqueue/cmd/job/internal/svc"
	"looklook/app/mqueue/cmd/job/jobtype"
	"looklook/app/order/model"
	"looklook/app/usercenter/cmd/rpc/usercenter"
	usercenterModel "looklook/app/usercenter/model"
	"looklook/common/globalkey"
	"looklook/common/tool"
	"looklook/common/wxminisub"
	"looklook/common/xerr"
	"time"
)

var ErrPaySuccessNotifyFail = xerr.NewErrMsg("pay success notify user fail")


// PaySuccessNotifyUserHandler pay success notify user
type PaySuccessNotifyUserHandler struct {
	svcCtx *svc.ServiceContext
}

func NewPaySuccessNotifyUserHandler(svcCtx *svc.ServiceContext) *PaySuccessNotifyUserHandler {
	return &PaySuccessNotifyUserHandler{
		svcCtx:svcCtx,
	}
}

func (l *PaySuccessNotifyUserHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {

	var p jobtype.PaySuccessNotifyUserPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return errors.Wrapf(ErrPaySuccessNotifyFail, "PaySuccessNotifyUserHandler payload err:%v, payLoad:%+v", err, t.Payload())
	}

	// 1、get user openid
	usercenterResp, err := l.svcCtx.UsercenterRpc.GetUserAuthByUserId(ctx, &usercenter.GetUserAuthByUserIdReq{
		UserId:   p.Order.UserId,
		AuthType: usercenterModel.UserAuthTypeSmallWX,
	})
	if err != nil {
		return errors.Wrapf(ErrPaySuccessNotifyFail,"pay success notify user fail, rpc get user err:%v , orderSn:%s , userId:%d",err,p.Order.Sn,p.Order.UserId)
	}
	if usercenterResp.UserAuth == nil || len(usercenterResp.UserAuth.AuthKey) == 0 {
		return errors.Wrapf(ErrPaySuccessNotifyFail,"pay success notify user , user no exists err:%v , orderSn:%s , userId:%d",err,p.Order.Sn,p.Order.UserId)
	}
	openId := usercenterResp.UserAuth.AuthKey


	// 2、send notify
	msgs := l.getData(ctx,p.Order,openId)
	for _, msg := range msgs  {
		l.SendWxMini(ctx,msg)
	}

	return nil
}

// get send data
func (l *PaySuccessNotifyUserHandler) getData(_ context.Context,order *model.HomestayOrder,openId string)[]*subscribe.Message{

	return []*subscribe.Message{
		{
			ToUser:    openId,
			TemplateID: wxminisub.OrderPaySuccessTemplateID,
			Data: map[string]*subscribe.DataItem{
				"character_string6": {Value: order.Sn},
				"thing1":            {Value: order.Title},
				"amount2":           {Value:fmt.Sprintf("%.2f", tool.Fen2Yuan(order.OrderTotalPrice))},
				"time4":             {Value:carbon.CreateFromTimestamp(order.LiveStartDate.Unix()).Format(globalkey.DateTimeFormatTplStandardDate)},
				"time5":             {Value:carbon.CreateFromTimestamp(order.LiveEndDate.Unix()).Format(globalkey.DateTimeFormatTplStandardDate)},
			},
		},
		{
			ToUser:    openId,
			TemplateID: wxminisub.OrderPaySuccessLiveKnowTemplateID,
			Data: map[string]*subscribe.DataItem{
				"date2":             {Value:carbon.CreateFromTimestamp(order.LiveStartDate.Unix()).Format(globalkey.DateTimeFormatTplStandardDate)},
				"date3":             {Value:carbon.CreateFromTimestamp(order.LiveEndDate.Unix()).Format(globalkey.DateTimeFormatTplStandardDate)} ,
				"character_string4": {Value:order.TradeCode} ,
				"thing1":            {Value:"请不要将验证码告知商家以外人员，以防上当"} ,
			},
		},
	}
}


// SendWxMini send to wechat mini
func (l *PaySuccessNotifyUserHandler) SendWxMini(ctx context.Context,msg *subscribe.Message)  {

	if l.svcCtx.Config.Mode != service.ProMode{
		msg.MiniprogramState = "developer"
	} else {
		msg.MiniprogramState = "formal"
	}

	var maxRetryNum int64 = 5
	var retryNum int64

	//Prevent user slowdown, delays, retries
	for {
		time.Sleep(time.Duration(1) * time.Second)

		err := l.svcCtx.MiniProgram.GetSubscribe().Send(msg)
		if err != nil {
			if retryNum > maxRetryNum {
				logx.WithContext(ctx).Errorf("Payment successful send wechat mini subscription message failed retryNum ： %d , err:%v, msg ： %+v ", retryNum,err, msg)
				return
			}
			retryNum++
			continue
		}

		return
	}
}
