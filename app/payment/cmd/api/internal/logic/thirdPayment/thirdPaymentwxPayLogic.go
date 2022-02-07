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

var ErrWxPayError = xerr.NewErrMsg("微信支付失败")

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

	var totalPrice int64   // 当前订单支付总金额(分)
	var description string // 当前支付描述.

	switch req.ServiceType {
	case model.ThirdPaymentServiceTypeHomestayOrder:

		homestayTotalPrice, homestayDescription, err := l.getPayHomestayPriceDescription(req.OrderSn)
		if err != nil {
			return nil, errors.Wrapf(ErrWxPayError, "getPayHomestayPriceDescription err : %v req: %+v", err, req)
		}
		totalPrice = homestayTotalPrice
		description = homestayDescription

	default:
		return nil, errors.Wrapf(xerr.NewErrMsg("不支持此业务类型支付"), "不支持此业务类型支付 req: %+v", req)
	}

	// 创建微信预处理订单
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

// 获取支付民宿当前订单的价格以及描述信息
func (l *ThirdPaymentwxPayLogic) createWxPrePayOrder(serviceType, orderSn string, totalPrice int64, description string) (*jsapi.PrepayWithRequestPaymentResponse, error) {

	// 1、获取用户openId
	userId := ctxdata.GetUidFromCtx(l.ctx)
	userResp, err := l.svcCtx.UsercenterRpc.GetUserAuthByUserId(l.ctx, &usercenter.GetUserAuthByUserIdReq{
		UserId:   userId,
		AuthType: usercenterModel.UserAuthTypeSmallWX,
	})
	if err != nil {
		return nil, errors.Wrapf(ErrWxPayError, "获取用户微信openid err : %v , userId: %d , orderSn:%s", err, userId, orderSn)
	}
	if userResp.UserAuth == nil || userResp.UserAuth.Id == 0 {
		return nil, errors.Wrapf(xerr.NewErrMsg("获取openid失败，请先微信授权在支付"), "获取用户微信openid不存在  userId: %d , orderSn:%s", userId, orderSn)
	}
	openId := userResp.UserAuth.AuthKey

	// 2、创建本地流水记录
	createPaymentResp, err := l.svcCtx.PaymentRpc.CreatePayment(l.ctx, &payment.CreatePaymentReq{
		UserId:      userId,
		PayModel:    model.ThirdPaymentPayModelWechatPay,
		PayTotal:    totalPrice,
		OrderSn:     orderSn,
		ServiceType: serviceType,
	})
	if err != nil || createPaymentResp.Sn == "" {
		return nil, errors.Wrapf(ErrWxPayError,
			"创建本地流水失败: err: %v , userId: %d,totalPrice: %d , orderSn: %s",
			err, userId, totalPrice, orderSn)
	}

	// 3、创建微信预处订单.

	wxPayClient, err := svc.NewWxPayClientV3(l.svcCtx.Config)
	if err != nil {
		return nil, err
	}
	jsApiSvc := jsapi.JsapiApiService{Client: wxPayClient}

	// 得到prepay_id，以及调起支付所需的参数和签名
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
		return nil, errors.Wrapf(ErrWxPayError, "发起微信支付预订单失败 err : %v , userId: %d , orderSn:%s", err, userId, orderSn)
	}

	return resp, nil

}

// 获取支付民宿当前订单的价格以及描述信息
func (l *ThirdPaymentwxPayLogic) getPayHomestayPriceDescription(orderSn string) (int64, string, error) {

	description := "民宿支付"

	// 获取用户openid
	resp, err := l.svcCtx.OrderRpc.HomestayOrderDetail(l.ctx, &order.HomestayOrderDetailReq{
		Sn: orderSn,
	})
	if err != nil {
		return 0, description, errors.Wrapf(ErrWxPayError,
			"OrderRpc.HomestayOrderDetail err: %v, orderSn: %s", err, orderSn)
	}
	if resp.HomestayOrder == nil || resp.HomestayOrder.Id == 0 {
		return 0, description, errors.Wrapf(xerr.NewErrMsg("订单不存在"), "微信支付订单不存在 orderSn : %s", orderSn)
	}

	return resp.HomestayOrder.OrderTotalPrice, description, nil
}
