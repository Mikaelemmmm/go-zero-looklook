package ctxdata

import (
	"context"
	"encoding/json"
)

// CtxKeyJwtUserId get uid from ctx
var CtxKeyJwtUserId = "jwtUserId"

func GetUidFromCtx(ctx context.Context) int64 {
	uid ,_ := ctx.Value(CtxKeyJwtUserId).(json.Number).Int64()
	return uid
}
