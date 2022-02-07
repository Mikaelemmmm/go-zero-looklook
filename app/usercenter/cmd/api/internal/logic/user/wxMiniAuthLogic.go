package user

import (
	"context"
	"fmt"

	"looklook/app/identity/cmd/rpc/identity"
	"looklook/app/usercenter/cmd/api/internal/svc"
	"looklook/app/usercenter/cmd/api/internal/types"
	"looklook/app/usercenter/cmd/rpc/usercenter"
	usercenterModel "looklook/app/usercenter/model"
	"looklook/common/xerr"

	"github.com/pkg/errors"
	wechat "github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	miniConfig "github.com/silenceper/wechat/v2/miniprogram/config"
	"github.com/zeromicro/go-zero/core/logx"
)

//error信息.
var ErrWxMiniAuthFailError = xerr.NewErrMsg("微信授权失败")

type WxMiniAuthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWxMiniAuthLogic(ctx context.Context, svcCtx *svc.ServiceContext) WxMiniAuthLogic {
	return WxMiniAuthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 微信小程序授权
func (l *WxMiniAuthLogic) WxMiniAuth(req types.WXMiniAuthReq) (*types.WXMiniAuthResp, error) {

	//1、授权
	miniprogram := wechat.NewWechat().GetMiniProgram(&miniConfig.Config{
		AppID:     l.svcCtx.Config.WxMiniConf.AppId,
		AppSecret: l.svcCtx.Config.WxMiniConf.Secret,
		Cache:     cache.NewMemory(),
	})
	authResult, err := miniprogram.GetAuth().Code2Session(req.Code)
	if err != nil || authResult.ErrCode != 0 || authResult.OpenID == "" {
		return nil, errors.Wrapf(ErrWxMiniAuthFailError, "发起授权请求失败 err : %v , code : %s  , authResult : %+v", err, req.Code, authResult)
	}
	//2、解析小程序返回数据
	userData, err := miniprogram.GetEncryptor().Decrypt(authResult.SessionKey, req.EncryptedData, req.IV)
	if err != nil {
		return nil, errors.Wrapf(ErrWxMiniAuthFailError, "解析数据失败 req : %+v , err: %v , authResult:%+v ", req, err, authResult)
	}

	//3、绑定用户 or 登陆.
	var userId int64
	rpcRsp, err := l.svcCtx.UsercenterRpc.GetUserAuthByAuthKey(l.ctx, &usercenter.GetUserAuthByAuthKeyReq{
		AuthType: usercenterModel.UserAuthTypeSmallWX,
		AuthKey:  authResult.OpenID,
	})
	if err != nil {
		return nil, errors.Wrapf(ErrWxMiniAuthFailError, "rpc call userAuthByAuthKey err : %v , authResult : %+v", err, authResult)
	}
	if rpcRsp.UserAuth == nil || rpcRsp.UserAuth.Id == 0 {

		//绑定用户.

		//小程序解密返回的数据都如下
		mobile := userData.PhoneNumber
		nickName := fmt.Sprintf("LookLook%s", mobile[7:])
		registerRsp, err := l.svcCtx.UsercenterRpc.Register(l.ctx, &usercenter.RegisterReq{
			AuthKey:  authResult.OpenID,
			AuthType: usercenterModel.UserAuthTypeSmallWX,
			Mobile:   mobile,
			Nickname: nickName,
		})
		if err != nil {
			return nil, errors.Wrapf(ErrWxMiniAuthFailError, "UsercenterRpc.Register err :%v, authResult : %+v", err, authResult)
		}

		return &types.WXMiniAuthResp{
			AccessToken:  registerRsp.AccessToken,
			AccessExpire: registerRsp.AccessExpire,
			RefreshAfter: registerRsp.RefreshAfter,
		}, nil

	} else {
		//登陆 直接授权返回token

		userId = rpcRsp.UserAuth.UserId

		tokenResp, err := l.svcCtx.IdentityRpc.GenerateToken(l.ctx, &identity.GenerateTokenReq{
			UserId: userId,
		})
		if err != nil {
			return nil, errors.Wrapf(ErrWxMiniAuthFailError, "IdentityRpc.GenerateToken err :%v, userId : %d", err, userId)
		}
		return &types.WXMiniAuthResp{
			AccessToken:  tokenResp.AccessToken,
			AccessExpire: tokenResp.AccessExpire,
			RefreshAfter: tokenResp.RefreshAfter,
		}, nil
	}
}
