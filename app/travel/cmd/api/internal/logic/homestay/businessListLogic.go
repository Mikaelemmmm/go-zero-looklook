package homestay

import (
	"context"
	"github.com/Masterminds/squirrel"

	"looklook/app/travel/cmd/api/internal/svc"
	"looklook/app/travel/cmd/api/internal/types"
	"looklook/pkg/tool"
	"looklook/pkg/xerr"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type BusinessListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBusinessListLogic(ctx context.Context, svcCtx *svc.ServiceContext) BusinessListLogic {
	return BusinessListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BusinessListLogic) BusinessList(req types.BusinessListReq) (*types.BusinessListResp, error) {

	whereBuilder := l.svcCtx.HomestayModel.SelectBuilder().Where(squirrel.Eq{"homestay_business_id": req.HomestayBusinessId})
	list, err := l.svcCtx.HomestayModel.FindPageListByIdDESC(l.ctx, whereBuilder, req.LastId, req.PageSize)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "HomestayBusinessId: %d ,err : %v", req.HomestayBusinessId, err)
	}

	var resp []types.Homestay
	if len(list) > 0 {
		// 【!!notice!!】Why not use copier to make a copy of the whole list here?
		// 【!!重要!!】这里为什么不使用copier去对整个list进行拷贝？

		// answer : copier This library is essentially the use of reflection implementation, in our online practice, the copy of large slices will take up a lot of cpu, serious performance consumption, if you can manually assign the value as much as possible manually, would like to use the copier is highly recommended only copy a single object is not a great impact
		// 答：copier 这个库本质上是使用反射实现的，在我们线上实践中，对大切片拷贝会占用大量的cpu，严重消耗性能，如果能手动赋值尽量手动，想使用copier强烈建议只拷贝单个对象影响不是很大
		for _, homestay := range list {

			var typeHomestay types.Homestay
			_ = copier.Copy(&typeHomestay, homestay)

			typeHomestay.FoodPrice = tool.Fen2Yuan(homestay.FoodPrice)
			typeHomestay.HomestayPrice = tool.Fen2Yuan(homestay.HomestayPrice)
			typeHomestay.MarketHomestayPrice = tool.Fen2Yuan(homestay.MarketHomestayPrice)

			resp = append(resp, typeHomestay)
		}
	}

	return &types.BusinessListResp{
		List: resp,
	}, nil
}
