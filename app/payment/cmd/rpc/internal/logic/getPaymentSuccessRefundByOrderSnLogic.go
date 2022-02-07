package logic

import (
	"context"

	"looklook/app/payment/cmd/rpc/internal/svc"
	"looklook/app/payment/cmd/rpc/pb"
	"looklook/app/payment/model"
	"looklook/common/xerr"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetPaymentSuccessRefundByOrderSnLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPaymentSuccessRefundByOrderSnLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaymentSuccessRefundByOrderSnLogic {
	return &GetPaymentSuccessRefundByOrderSnLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 根据订单sn查询流水记..
func (l *GetPaymentSuccessRefundByOrderSnLogic) GetPaymentSuccessRefundByOrderSn(in *pb.GetPaymentSuccessRefundByOrderSnReq) (*pb.GetPaymentSuccessRefundByOrderSnResp, error) {

	thirdPayment, err := l.svcCtx.ThirdPaymentModel.FindOnePaySucessOrRefundByOrderSn(in.OrderSn)
	if err != nil && err != model.ErrNotFound {
		return nil, errors.Wrapf(xerr.NewErrMsg("获取支付流水信息失败"), "获取支付流水信息失败 err : %v , orderSn:%s", err, in.OrderSn)
	}

	var resp pb.PaymentDetail
	if thirdPayment != nil {

		_ = copier.Copy(&resp, thirdPayment)
		resp.CreateTime = thirdPayment.CreateTime.Unix()
		resp.UpdateTime = thirdPayment.UpdateTime.Unix()
		resp.PayTime = thirdPayment.PayTime.Unix()

	}

	return &pb.GetPaymentSuccessRefundByOrderSnResp{
		PaymentDetail: &resp,
	}, nil
}
