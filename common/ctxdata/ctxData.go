package ctxdata

import (
	"context"
)

//从ctx获取uid
var CtxKeyJwtUserId = "jwtUserId"

func GetUidFromCtx(ctx context.Context) int64 {
	uid, _ := ctx.Value(CtxKeyJwtUserId).(int64)
	return uid
}
