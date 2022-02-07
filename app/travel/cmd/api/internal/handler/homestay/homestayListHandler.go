package homestay

import (
	"net/http"

	"looklook/app/travel/cmd/api/internal/logic/homestay"
	"looklook/app/travel/cmd/api/internal/svc"
	"looklook/app/travel/cmd/api/internal/types"
	"looklook/common/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func HomestayListHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.HomestayListReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := homestay.NewHomestayListLogic(r.Context(), ctx)
		resp, err := l.HomestayList(req)
		result.HttpResult(r, w, resp, err)
	}
}
