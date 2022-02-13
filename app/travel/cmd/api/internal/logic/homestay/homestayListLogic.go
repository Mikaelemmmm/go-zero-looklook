package homestay

import (
	"context"

	"looklook/app/travel/cmd/api/internal/svc"
	"looklook/app/travel/cmd/api/internal/types"
	"looklook/app/travel/model"
	"looklook/common/tool"
	"looklook/common/xerr"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/mr"
)

type HomestayListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

var ErrHomestayListError = xerr.NewErrMsg("获取民宿列表失败")

func NewHomestayListLogic(ctx context.Context, svcCtx *svc.ServiceContext) HomestayListLogic {
	return HomestayListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 获取民宿列表
func (l *HomestayListLogic) HomestayList(req types.HomestayListReq) (*types.HomestayListResp, error) {

	switch req.RowType {
	default:
		return l.getActivityList(req)
	}
}

// 活动民宿
func (l *HomestayListLogic) getActivityList(req types.HomestayListReq) (*types.HomestayListResp, error) {

	// 获取活动数据id集合.
	homestayIds, err := l.svcCtx.HomestayActivityModel.FindPageByRowTypeStatus(req.LastId, req.PageSize, req.RowType, model.HomestayActivityUpStatus)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "rowType: %s ,err : %v", req.RowType, err)
	}

	var resp []types.Homestay
	if len(homestayIds) > 0 { // mr从缓存中捞数据
		mr.MapReduceVoid(func(source chan<- interface{}) {
			for _, id := range homestayIds {
				source <- id
			}
		}, func(item interface{}, writer mr.Writer, cancel func(error)) {
			id := item.(int64)

			homestay, err := l.svcCtx.HomestayModel.FindOne(id)
			if err != nil && err != model.ErrNotFound {
				// 列表数据不返回错误，记录日志即可.
				logx.WithContext(l.ctx).Errorf("ActivityHomestayListLogic ActivityHomestayList 获取活动数据失败 id : %d ,err : %v", id, err)
				return
			}
			writer.Write(homestay)
		}, func(pipe <-chan interface{}, cancel func(error)) {

			for item := range pipe {
				homestay := item.(*model.Homestay)
				var tyHomestay types.Homestay
				_ = copier.Copy(&tyHomestay, homestay)

				tyHomestay.FoodPrice = tool.Fen2Yuan(homestay.FoodPrice)
				tyHomestay.HomestayPrice = tool.Fen2Yuan(homestay.HomestayPrice)
				tyHomestay.MarketHomestayPrice = tool.Fen2Yuan(homestay.MarketHomestayPrice)

				resp = append(resp, tyHomestay)
			}
		})
	}

	return &types.HomestayListResp{
		List: resp,
	}, nil
}
