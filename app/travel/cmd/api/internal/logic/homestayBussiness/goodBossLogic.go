package homestayBussiness

import (
	"context"
	"github.com/Masterminds/squirrel"

	"looklook/app/travel/cmd/api/internal/svc"
	"looklook/app/travel/cmd/api/internal/types"
	"looklook/app/travel/model"
	"looklook/app/usercenter/cmd/rpc/usercenter"
	"looklook/common/xerr"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/mr"
)

type GoodBossLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGoodBossLogic(ctx context.Context, svcCtx *svc.ServiceContext) GoodBossLogic {
	return GoodBossLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GoodBossLogic) GoodBoss(req types.GoodBossReq) (*types.GoodBossResp, error) {

	// 获取10个最佳房东.

	whereBuilder := l.svcCtx.HomestayActivityModel.RowBuilder().Where(squirrel.Eq{
		"row_type":  model.HomestayActivityGoodBusiType,
		"row_status" : model.HomestayActivityUpStatus,
	})
	homestayActivityList, err := l.svcCtx.HomestayActivityModel.FindPageListByPage(whereBuilder,0, 10,"data_id desc")
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "获取10个最佳房东. rowType: %s ,err : %v", model.HomestayActivityGoodBusiType, err)
	}

	var resp []types.HomestayBusinessBoss
	if len(homestayActivityList) > 0 {

		mr.MapReduceVoid(func(source chan<- interface{}) {
			for _, homestayActivity := range homestayActivityList {
				source <- homestayActivity.DataId
			}
		}, func(item interface{}, writer mr.Writer, cancel func(error)) {
			id := item.(int64)

			userResp, err := l.svcCtx.UsercenterRpc.GetUserInfo(l.ctx, &usercenter.GetUserInfoReq{
				Id: id,
			})
			if err != nil {
				logx.WithContext(l.ctx).Errorf("GoodListLogic GoodList最佳房东获取房东信息失败 userId : %d ,err:%v", id, err)
				return
			}
			if userResp.User != nil && userResp.User.Id > 0 {
				writer.Write(userResp.User)
			}
		}, func(pipe <-chan interface{}, cancel func(error)) {

			for item := range pipe {
				var typesHomestayBusiness types.HomestayBusinessBoss
				_ = copier.Copy(&typesHomestayBusiness, item)

				// 计算star todo
				resp = append(resp, typesHomestayBusiness)
			}
		})
	}

	return &types.GoodBossResp{
		List: resp,
	}, nil
}
