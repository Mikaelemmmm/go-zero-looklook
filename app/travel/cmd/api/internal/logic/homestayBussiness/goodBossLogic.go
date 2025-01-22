package homestayBussiness

import (
	"context"
	"github.com/Masterminds/squirrel"

	"looklook/app/travel/cmd/api/internal/svc"
	"looklook/app/travel/cmd/api/internal/types"
	"looklook/app/travel/model"
	"looklook/app/usercenter/cmd/rpc/usercenter"
	"looklook/pkg/xerr"

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

	whereBuilder := l.svcCtx.HomestayActivityModel.SelectBuilder().Where(squirrel.Eq{
		"row_type":   model.HomestayActivityGoodBusiType,
		"row_status": model.HomestayActivityUpStatus,
	})
	homestayActivityList, err := l.svcCtx.HomestayActivityModel.FindPageListByPage(l.ctx, whereBuilder, 0, 10, "data_id desc")
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "get GoodBoss db err. rowType: %s ,err : %v", model.HomestayActivityGoodBusiType, err)
	}

	var resp []types.HomestayBusinessBoss
	if len(homestayActivityList) > 0 {

		mr.MapReduceVoid(func(source chan<- interface{}) {
			for _, homestayActivity := range homestayActivityList {
				source <- homestayActivity.DataId
			}
		}, func(item interface{}, writer mr.Writer[*usercenter.User], cancel func(error)) {
			id := item.(int64)

			userResp, err := l.svcCtx.UsercenterRpc.GetUserInfo(l.ctx, &usercenter.GetUserInfoReq{
				Id: id,
			})
			if err != nil {
				logx.WithContext(l.ctx).Errorf("GoodListLogic GoodList fail userId : %d ,err:%v", id, err)
				return
			}
			if userResp.User != nil && userResp.User.Id > 0 {
				writer.Write(userResp.User)
			}
		}, func(pipe <-chan *usercenter.User, cancel func(error)) {

			// 【!!notice!!】Why not use copier to make a copy of the whole list here?
			// 【!!重要!!】这里为什么不使用copier去对整个list进行拷贝？

			// answer : copier This library is essentially the use of reflection implementation, in our online practice, the copy of large slices will take up a lot of cpu, serious performance consumption, if you can manually assign the value as much as possible manually, would like to use the copier is highly recommended only copy a single object is not a great impact
			// 答：copier 这个库本质上是使用反射实现的，在我们线上实践中，对大切片拷贝会占用大量的cpu，严重消耗性能，如果能手动赋值尽量手动，想使用copier强烈建议只拷贝单个对象影响不是很大
			for item := range pipe {
				var typesHomestayBusiness types.HomestayBusinessBoss
				_ = copier.Copy(&typesHomestayBusiness, item)

				// compute star todo
				resp = append(resp, typesHomestayBusiness)
			}
		})
	}

	return &types.GoodBossResp{
		List: resp,
	}, nil
}
