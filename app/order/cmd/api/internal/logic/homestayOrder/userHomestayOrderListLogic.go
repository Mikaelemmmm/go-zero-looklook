package homestayOrder

import (
	"context"

	"looklook/app/order/cmd/api/internal/svc"
	"looklook/app/order/cmd/api/internal/types"
	"looklook/app/order/cmd/rpc/order"
	"looklook/pkg/ctxdata"
	"looklook/pkg/tool"
	"looklook/pkg/xerr"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type UserHomestayOrderListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserHomestayOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) UserHomestayOrderListLogic {
	return UserHomestayOrderListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserHomestayOrderListLogic) UserHomestayOrderList(req types.UserHomestayOrderListReq) (*types.UserHomestayOrderListResp, error) {

	userId := ctxdata.GetUidFromCtx(l.ctx) //get login user id

	resp, err := l.svcCtx.OrderRpc.UserHomestayOrderList(l.ctx, &order.UserHomestayOrderListReq{
		UserId:      userId,
		TraderState: req.TradeState,
		PageSize:    req.PageSize,
		LastId:      req.LastId,
	})
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrMsg("Failed to get user homestay order list"), "Failed to get user homestay order list err : %v ,req:%+v", err, req)
	}

	var typesUserHomestayOrderList []types.UserHomestayOrderListView

	if len(resp.List) > 0 {

		// 【!!notice!!】Why not use copier to make a copy of the whole list here?
		// 【!!重要!!】这里为什么不使用copier去对整个list进行拷贝？

		// answer : copier This library is essentially the use of reflection implementation, in our online practice, the copy of large slices will take up a lot of cpu, serious performance consumption, if you can manually assign the value as much as possible manually, would like to use the copier is highly recommended only copy a single object is not a great impact
		// 答：copier 这个库本质上是使用反射实现的，在我们线上实践中，对大切片拷贝会占用大量的cpu，严重消耗性能，如果能手动赋值尽量手动，想使用copier强烈建议只拷贝单个对象影响不是很大
		for _, homestayOrder := range resp.List {

			var typeHomestayOrder types.UserHomestayOrderListView
			_ = copier.Copy(&typeHomestayOrder, homestayOrder)

			typeHomestayOrder.OrderTotalPrice = tool.Fen2Yuan(homestayOrder.OrderTotalPrice)

			typesUserHomestayOrderList = append(typesUserHomestayOrderList, typeHomestayOrder)
		}
	}

	return &types.UserHomestayOrderListResp{
		List: typesUserHomestayOrderList,
	}, nil
}
