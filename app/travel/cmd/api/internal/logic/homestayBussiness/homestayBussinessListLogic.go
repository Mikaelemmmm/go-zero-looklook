package homestayBussiness

import (
	"context"

	"looklook/app/travel/cmd/api/internal/svc"
	"looklook/app/travel/cmd/api/internal/types"
	"looklook/pkg/xerr"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type HomestayBussinessListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHomestayBussinessListLogic(ctx context.Context, svcCtx *svc.ServiceContext) HomestayBussinessListLogic {
	return HomestayBussinessListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HomestayBussinessListLogic) HomestayBussinessList(req types.HomestayBussinessListReq) (*types.HomestayBussinessListResp, error) {

	whereBuilder := l.svcCtx.HomestayBusinessModel.SelectBuilder()
	list, err := l.svcCtx.HomestayBusinessModel.FindPageListByIdDESC(l.ctx, whereBuilder, req.LastId, req.PageSize)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "HomestayBussinessList FindPageListByIdDESC db fail ,  req : %+v , err:%v", req, err)
	}

	var resp []types.HomestayBusinessListInfo

	// 【!!notice!!】Why not use copier to make a copy of the whole list here?
	// 【!!重要!!】这里为什么不使用copier去对整个list进行拷贝？

	// answer : copier This library is essentially the use of reflection implementation, in our online practice, the copy of large slices will take up a lot of cpu, serious performance consumption, if you can manually assign the value as much as possible manually, would like to use the copier is highly recommended only copy a single object is not a great impact
	// 答：copier 这个库本质上是使用反射实现的，在我们线上实践中，对大切片拷贝会占用大量的cpu，严重消耗性能，如果能手动赋值尽量手动，想使用copier强烈建议只拷贝单个对象影响不是很大
	if len(list) > 0 {
		for _, item := range list {
			var typeHomestayBusinessListInfo types.HomestayBusinessListInfo
			_ = copier.Copy(&typeHomestayBusinessListInfo, item)

			resp = append(resp, typeHomestayBusinessListInfo)
		}
	}

	return &types.HomestayBussinessListResp{
		List: resp,
	}, nil
}
