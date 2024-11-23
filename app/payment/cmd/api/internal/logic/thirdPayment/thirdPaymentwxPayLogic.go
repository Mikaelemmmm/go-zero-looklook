package thirdPayment

import (
	"context"

	"looklook/app/order/cmd/rpc/order"
	"looklook/app/payment/cmd/api/internal/svc"
	"looklook/app/payment/cmd/api/internal/types"
	"looklook/app/payment/cmd/rpc/payment"
	"looklook/app/payment/model"
	"looklook/app/usercenter/cmd/rpc/usercenter"
	usercenterModel "looklook/app/usercenter/model"
	"looklook/common/ctxdata"
	"looklook/common/xerr"

	"github.com/pkg/errors"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/zeromicro/go-zero/core/logx"
)

var ErrWxPayError = xerr.NewErrMsg("wechat pay fail")

type ThirdPaymentwxPayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewThirdPaymentwxPayLogic(ctx context.Context, svcCtx *svc.ServiceContext) ThirdPaymentwxPayLogic {
	return ThirdPaymentwxPayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ThirdPaymentwxPayLogic) ThirdPaymentwxPay(req types.ThirdPaymentWxPayReq) (*types.ThirdPaymentWxPayResp, error) {

	var totalPrice int64   // Total amount paid for current order(cent)
	var description string // Current Payment Description.

	switch req.ServiceType {
	case model.ThirdPaymentServiceTypeHomestayOrder:

		homestayTotalPrice, homestayDescription, err := l.getPayHomestayPriceDescription(req.OrderSn)
		if err != nil {
			return nil, errors.Wrapf(ErrWxPayError, "getPayHomestayPriceDescription err : %v req: %+v", err, req)
		}
		totalPrice = homestayTotalPrice
		description = homestayDescription

	default:
		return nil, errors.Wrapf(xerr.NewErrMsg("Payment for this business type is not supported"), "Payment for this business type is not supported req: %+v", req)
	}

	// Create WechatPay pre-processing orders
	wechatPrepayRsp, err := l.createWxPrePayOrder(req.ServiceType, req.OrderSn, totalPrice, description)
	if err != nil {
		return nil, err
	}

	return &types.ThirdPaymentWxPayResp{
		Appid:     l.svcCtx.Config.WxMiniConf.AppId,
		NonceStr:  *wechatPrepayRsp.NonceStr,
		PaySign:   *wechatPrepayRsp.PaySign,
		Package:   *wechatPrepayRsp.Package,
		Timestamp: *wechatPrepayRsp.TimeStamp,
		SignType:  *wechatPrepayRsp.SignType,
	}, nil
}

// Get the price and description information of the current order of the paid B&B
func (l *ThirdPaymentwxPayLogic) createWxPrePayOrder(serviceType, orderSn string, totalPrice int64, description string) (*jsapi.PrepayWithRequestPaymentResponse, error) {

	// 1、get user openId
	userId := ctxdata.GetUidFromCtx(l.ctx)
	userResp, err := l.svcCtx.UsercenterRpc.GetUserAuthByUserId(l.ctx, &usercenter.GetUserAuthByUserIdReq{
		UserId:   userId,
		AuthType: usercenterModel.UserAuthTypeSmallWX,
	})
	if err != nil {
		return nil, errors.Wrapf(ErrWxPayError, "Get user wechat openid err : %v , userId: %d , orderSn:%s", err, userId, orderSn)
	}
	if userResp.UserAuth == nil || userResp.UserAuth.Id == 0 {
		return nil, errors.Wrapf(xerr.NewErrMsg("Get user wechat openid fail，Please pay before authorization by weChat"), "Get user WeChat openid does not exist  userId: %d , orderSn:%s", userId, orderSn)
	}
	openId := userResp.UserAuth.AuthKey

	// 2、create local third payment record
	createPaymentResp, err := l.svcCtx.PaymentRpc.CreatePayment(l.ctx, &payment.CreatePaymentReq{
		UserId:      userId,
		PayModel:    model.ThirdPaymentPayModelWechatPay,
		PayTotal:    totalPrice,
		OrderSn:     orderSn,
		ServiceType: serviceType,
	})
	if err != nil || createPaymentResp.Sn == "" {
		return nil, errors.Wrapf(ErrWxPayError,
			"create local third payment record fail : err: %v , userId: %d,totalPrice: %d , orderSn: %s",
			err, userId, totalPrice, orderSn)
	}

	// 3、create wechat pay pre pay order

	wxPayClient, err := svc.NewWxPayClientV3(l.svcCtx.Config)
	if err != nil {
		return nil, err
	}
	jsApiSvc := jsapi.JsapiApiService{Client: wxPayClient}

	// Get the prepay_id, as well as the parameters and signatures needed to invoke the payment
	resp, _, err := jsApiSvc.PrepayWithRequestPayment(l.ctx,
		jsapi.PrepayRequest{
			Appid:       core.String(l.svcCtx.Config.WxMiniConf.AppId),
			Mchid:       core.String(l.svcCtx.Config.WxPayConf.MchId),
			Description: core.String(description),
			OutTradeNo:  core.String(createPaymentResp.Sn),
			Attach:      core.String(description),
			NotifyUrl:   core.String(l.svcCtx.Config.WxPayConf.NotifyUrl),
			Amount: &jsapi.Amount{
				Total: core.Int64(totalPrice),
			},
			Payer: &jsapi.Payer{
				Openid: core.String(openId),
			},
		},
	)
	if err != nil {
		return nil, errors.Wrapf(ErrWxPayError, "Failed to initiate WeChat payment pre-order err : %v , userId: %d , orderSn:%s", err, userId, orderSn)
	}

	return resp, nil

}

// Get the price and description information of the current order of the paid B&B
func (l *ThirdPaymentwxPayLogic) getPayHomestayPriceDescription(orderSn string) (int64, string, error) {

	description := "homestay pay"

	// get user openid
	resp, err := l.svcCtx.OrderRpc.HomestayOrderDetail(l.ctx, &order.HomestayOrderDetailReq{
		Sn: orderSn,
	})
	if err != nil {
		return 0, description, errors.Wrapf(ErrWxPayError,
			"OrderRpc.HomestayOrderDetail err: %v, orderSn: %s", err, orderSn)
	}
	if resp.HomestayOrder == nil || resp.HomestayOrder.Id == 0 {
		return 0, description, errors.Wrapf(xerr.NewErrMsg("order no exists"), "WeChat payment order does not exist orderSn : %s", orderSn)
	}

	return resp.HomestayOrder.OrderTotalPrice, description, nil
}
