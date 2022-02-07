package thirdPayment

import (
	"fmt"
	"net/http"

	"looklook/app/payment/cmd/api/internal/logic/thirdPayment"
	"looklook/app/payment/cmd/api/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

func ThirdPaymentWxPayCallbackHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		l := thirdPayment.NewThirdPaymentWxPayCallbackLogic(r.Context(), ctx)
		resp, err := l.ThirdPaymentWxPayCallback(w, r)

		if err != nil {
			logx.WithContext(r.Context()).Errorf("【API-ERR】 ThirdPaymentWxPayCallbackHandler : %+v ", err)
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		logx.Infof("ReturnCode : %s ", resp.ReturnCode)
		fmt.Fprint(w.(http.ResponseWriter), resp.ReturnCode)
	}
}
