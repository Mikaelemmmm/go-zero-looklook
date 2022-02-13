package verify

import (
	"context"
	"encoding/json"
	"github.com/zeromicro/go-zero/rest/token"
	"net/http"
	"strings"

	"looklook/app/identity/cmd/api/internal/svc"
	"looklook/app/identity/cmd/api/internal/types"
	"looklook/app/identity/cmd/rpc/identity"
	"looklook/common/ctxdata"
	"looklook/common/xerr"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type TokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

var ValidateTokenError = xerr.NewErrCode(xerr.TOKEN_EXPIRE_ERROR)

func NewTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) TokenLogic {
	return TokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TokenLogic) Token(req types.VerifyTokenReq, r *http.Request) (*types.VerifyTokenResp, error) {

	authorization := r.Header.Get("Authorization")
	realRequestPath := r.Header.Get("X-Original-Uri")

	if strings.Contains(realRequestPath, "?") {
		realRequestPath = strings.Split(realRequestPath, "?")[0]
	}

	var resultUserId int64
	if l.urlNoAuth(realRequestPath) {
		// 不需要登陆的页面.
		if len(authorization) > 0 { // 如果有传递token，就验证解析出来uid，没有token不验证..
			userId, err := l.isPass(r)
			if err != nil {
				logx.WithContext(l.ctx).Errorf("authorization:%s, realRequestPath:%s", authorization, realRequestPath)
				return nil, err
			}
			if userId == 0 {
				return nil, errors.Wrapf(ValidateTokenError, "urlIsAuth.false isPass userId  is 0 , authorization:%s, realRequestPath:%s", authorization, realRequestPath)
			}

			resultUserId = userId
		}
	} else {
		// 需要登陆的页面.
		userId, err := l.isPass(r)
		if err != nil {
			logx.WithContext(l.ctx).Errorf("authorization:%s, realRequestPath:%s", authorization, realRequestPath)
			return nil, err
		}
		if userId == 0 {
			return nil, errors.Wrapf(ValidateTokenError, "urlIsAuth.true isPass userId  is 0 , authorization: %s ,realRequestPath:%s", authorization, realRequestPath)
		}

		resultUserId = userId
	}

	return &types.VerifyTokenResp{
		UserId: resultUserId,
		Ok:     true,
	}, nil
}

// 获取token.
func (l *TokenLogic) getJwtToken(secretKey string, iat, seconds, userId int64) (string, error) {

	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims[ctxdata.CtxKeyJwtUserId] = userId
	t := jwt.New(jwt.SigningMethodHS256)
	t.Claims = claims
	return t.SignedString([]byte(secretKey))
}

// 当前url是否需要授权验证
func (l *TokenLogic) urlNoAuth(path string) bool {
	for _, val := range l.svcCtx.Config.NoAuthUrls {
		if val == path {
			return true
		}
	}
	return false
}

// 当前url是否需要授权验证.
func (l *TokenLogic) isPass(r *http.Request) (int64, error) {

	parser := token.NewTokenParser()
	tok, err := parser.ParseToken(r, l.svcCtx.Config.JwtAuth.AccessSecret, "")

	if err != nil {
		return 0, errors.Wrapf(ValidateTokenError, "JwtAuthLogic isPass  ParseToken err : %v", err)
	}
	if tok.Valid {
		claims, ok := tok.Claims.(jwt.MapClaims) // 解析token中对内容
		if ok {
			userId, _ := claims[ctxdata.CtxKeyJwtUserId].(json.Number).Int64() // 获取userId 并且到后端redis校验是否过期
			if userId <= 0 {
				return 0, errors.Wrapf(ValidateTokenError, "JwtAuthLogic.isPass invalid userId  tokRaw:%s , tokValid :%v ,userId:%d ", tok.Raw, tok.Valid, userId)
			}
			resp, err := l.svcCtx.IdentityRpc.ValidateToken(l.ctx, &identity.ValidateTokenReq{
				UserId: userId,
				Token:  tok.Raw,
			})
			if err != nil || !resp.Ok {
				return 0, errors.Wrapf(ValidateTokenError, "JwtAuthLogic.isPass IdentityRpc . ValidateToken err:%v ,resp:%+v , tokRaw:%s , tokValid : %v,userId:%d ", err, resp, tok.Raw, tok.Valid, userId)
			}
			return userId, nil
		} else {
			return 0, errors.Wrapf(ValidateTokenError, "tok.Claims is not ok ,tok.Claims ：%+v , claims : %+v , ok:%v ", tok.Claims, claims, ok)
		}
	}
	return 0, nil
}
