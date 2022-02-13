package homestayBussiness

import (
	"context"

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
	userIds, err := l.svcCtx.HomestayActivityModel.FindPageByRowTypeStatus(0, 10, model.HomestayActivityGoodBusiType, model.HomestayActivityUpStatus)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "req : %+v , err : %v ", req, err)
	}

	var resp []types.HomestayBusinessBoss
	if len(userIds) > 0 {

		mr.MapReduceVoid(func(source chan<- interface{}) {
			for _, id := range userIds {
				source <- id
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
