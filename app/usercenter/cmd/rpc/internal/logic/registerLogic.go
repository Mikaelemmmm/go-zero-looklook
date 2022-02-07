package logic

import (
	"context"
	"looklook/common/tool"

	"looklook/app/identity/cmd/rpc/identity"
	"looklook/app/usercenter/cmd/rpc/internal/svc"
	"looklook/app/usercenter/cmd/rpc/usercenter"
	"looklook/app/usercenter/model"
	"looklook/common/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var ErrUserAlreadyRegisterError = xerr.NewErrMsg("该用户已被注册")

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *usercenter.RegisterReq) (*usercenter.RegisterResp, error) {

	user, err := l.svcCtx.UserModel.FindOneByMobile(in.Mobile)
	if err != nil && err != model.ErrNotFound {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "mobile:%s,err:%v", in.Mobile, err)
	}
	if user != nil {
		return nil, errors.Wrapf(ErrUserAlreadyRegisterError, "用户已经存在 mobile:%s,err:%v", in.Mobile, err)
	}

	var userId int64
	if err := l.svcCtx.UserModel.Trans(func(session sqlx.Session) error {
		user := new(model.User)
		user.Mobile = in.Mobile
		if len(user.Nickname) == 0 {
			user.Nickname = tool.Krand(tool.KC_RAND_KIND_ALL, 8)
		}
		if len(user.Password) > 0 {
			user.Password = tool.Md5ByString(in.Password)
		}
		insertResult, err := l.svcCtx.UserModel.Insert(session, user)
		if err != nil {
			return errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "err:%v,user:%+v", err, user)
		}
		lastId, err := insertResult.LastInsertId()
		if err != nil {
			return errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "insertResult.LastInsertId err:%v,user:%+v", err, user)
		}
		userId = lastId

		userAuth := new(model.UserAuth)
		userAuth.UserId = lastId
		userAuth.AuthKey = in.AuthKey
		userAuth.AuthType = in.AuthType
		if _, err := l.svcCtx.UserAuthModel.Insert(session, userAuth); err != nil {
			return errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "err:%v,userAuth:%v", err, userAuth)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	//2、生成token.
	resp, err := l.svcCtx.IdentityRpc.GenerateToken(l.ctx, &identity.GenerateTokenReq{
		UserId: userId,
	})
	if err != nil {
		return nil, errors.Wrapf(ErrGenerateTokenError, "IdentityRpc.GenerateToken userId : %d , err:%+v", userId, err)
	}

	return &usercenter.RegisterResp{
		AccessToken:  resp.AccessToken,
		AccessExpire: resp.AccessExpire,
		RefreshAfter: resp.RefreshAfter,
	}, nil
}
