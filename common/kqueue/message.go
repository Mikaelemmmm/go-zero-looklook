//KqMessage
package kqueue

//第三方支付回调更改支付状态通知
type ThirdPaymentUpdatePayStatusNotifyMessage struct {
	PayStatus int64  `json:"payStatus"`
	OrderSn   string `json:"orderSn"`
}

