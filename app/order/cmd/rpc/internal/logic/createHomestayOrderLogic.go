package logic

import (
	"context"
	"strings"
	"time"

	"looklook/app/mqueue/cmd/rpc/mqueue"
	"looklook/app/order/cmd/rpc/internal/svc"
	"looklook/app/order/cmd/rpc/pb"
	"looklook/app/order/model"
	"looklook/app/travel/cmd/rpc/travel"
	"looklook/common/tool"
	"looklook/common/uniqueid"
	"looklook/common/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateHomestayOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateHomestayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateHomestayOrderLogic {
	return &CreateHomestayOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 民宿下订单.
func (l *CreateHomestayOrderLogic) CreateHomestayOrder(in *pb.CreateHomestayOrderReq) (*pb.CreateHomestayOrderResp, error) {

	//1、创建订单
	if in.LiveEndTime <= in.LiveStartTime {
		return nil, errors.Wrapf(xerr.NewErrMsg("至少要住一晚"), "民宿下订单 入住结束时间一定要大于开始时间 in : %+v", in)
	}

	resp, err := l.svcCtx.TravelRpc.HomestayDetail(l.ctx, &travel.HomestayDetailReq{
		Id: in.HomestayId,
	})
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrMsg("查询该记录失败"), "homestayId : %d , err : %v", in.HomestayId, err)
	}
	if resp.Homestay == nil {
		return nil, errors.Wrapf(xerr.NewErrMsg("该记录不存在"), "homestayId : %d ", in.HomestayId)
	}

	var cover string //获取封面..
	if len(resp.Homestay.Banner) > 0 {
		cover = strings.Split(resp.Homestay.Banner, ",")[0]
	}

	order := new(model.HomestayOrder)
	order.Sn = uniqueid.GenSn(uniqueid.SN_PREFIX_HOMESTAY_ORDER)
	order.UserId = in.UserId
	order.HomestayId = in.HomestayId
	order.Title = resp.Homestay.Title
	order.SubTitle = resp.Homestay.SubTitle
	order.Cover = cover
	order.Info = resp.Homestay.Info
	order.PeopleNum = resp.Homestay.PeopleNum
	order.RowType = resp.Homestay.RowType
	order.HomestayPrice = resp.Homestay.HomestayPrice
	order.MarketHomestayPrice = resp.Homestay.MarketHomestayPrice
	order.HomestayBusinessId = resp.Homestay.HomestayBusinessId
	order.HomestayUserId = resp.Homestay.UserId
	order.LivePeopleNum = in.LivePeopleNum
	order.TradeState = model.HomestayOrderTradeStateWaitPay
	order.TradeCode = tool.Krand(8, tool.KC_RAND_KIND_ALL)
	order.Remark = in.Remark
	order.FoodInfo = resp.Homestay.FoodInfo
	order.FoodPrice = resp.Homestay.FoodPrice
	order.LiveStartDate = time.Unix(in.LiveStartTime, 0)
	order.LiveEndDate = time.Unix(in.LiveEndTime, 0)

	liveDays := int64(order.LiveEndDate.Sub(order.LiveStartDate).Seconds() / 86400) //共住了几天

	order.HomestayTotalPrice = int64(resp.Homestay.HomestayPrice * liveDays) //计算民宿总价格
	if in.IsFood {
		order.NeedFood = model.HomestayOrderNeedFoodYes
		//计算餐食总价格.
		order.FoodTotalPrice = int64(resp.Homestay.FoodPrice * in.LivePeopleNum * liveDays)
	}

	order.OrderTotalPrice = order.HomestayTotalPrice + order.FoodTotalPrice //计算订单总价格.

	_, err = l.svcCtx.HomestayOrderModel.Insert(nil, order)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "下单数据库异常 order : %+v , err: %v", order, err)
	}

	//2、延迟关闭订单任务.
	_, _ = l.svcCtx.MqueueRpc.AqDeferHomestayOrderClose(l.ctx, &mqueue.AqDeferHomestayOrderCloseReq{
		Sn: order.Sn,
	})

	return &pb.CreateHomestayOrderResp{
		Sn: order.Sn,
	}, nil
}
