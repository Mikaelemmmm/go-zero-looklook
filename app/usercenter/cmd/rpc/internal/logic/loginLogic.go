package logic

import (
	"context"

	"looklook/app/identity/cmd/rpc/identity"
	"looklook/app/usercenter/cmd/rpc/internal/svc"
	"looklook/app/usercenter/cmd/rpc/usercenter"
	"looklook/app/usercenter/model"
	"looklook/common/tool"
	"looklook/common/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

var ErrGenerateTokenError = xerr.NewErrMsg("生成token失败")
var ErrUsernamePwdError = xerr.NewErrMsg("账号或密码不正确")

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *usercenter.LoginReq) (*usercenter.LoginResp, error) {

	var userId int64
	var err error
	switch in.AuthType {
	case model.UserAuthTypeSystem:
		userId, err = l.loginByMobile(in.AuthKey, in.Password)
	default:
		return nil, xerr.NewErrCode(xerr.SERVER_COMMON_ERROR)
	}
	if err != nil {
		return nil, err
	}

	//2、生成token
	resp, err := l.svcCtx.IdentityRpc.GenerateToken(l.ctx, &identity.GenerateTokenReq{
		UserId: userId,
	})
	if err != nil {
		return nil, errors.Wrapf(ErrGenerateTokenError, "IdentityRpc.GenerateToken userId : %d", userId)
	}

	return &usercenter.LoginResp{
		AccessToken:  resp.AccessToken,
		AccessExpire: resp.AccessExpire,
		RefreshAfter: resp.RefreshAfter,
	}, nil
}

func (l *LoginLogic) loginByMobile(mobile, password string) (int64, error) {

	user, err := l.svcCtx.UserModel.FindOneByMobile(mobile)
	if err != nil && err != model.ErrNotFound {
		return 0, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "根据手机号查询用户信息失败，mobile:%s,err:%v", mobile, err)
	}
	if user == nil {
		return 0, errors.Wrapf(ErrUserNoExistsError, "mobile:%s", mobile)
	}

	if !(tool.Md5ByString(password) == user.Password) {
		return 0, errors.Wrap(ErrUsernamePwdError, "密码匹配出错")
	}

	return user.Id, nil
}

func (l *LoginLogic) loginBySmallWx() error {
	return nil
}
