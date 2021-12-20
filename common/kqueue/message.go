//KqMessage
package kqueue

//第三方支付回调更改支付状态通知
type ThirdPaymentUpdatePayStatusNotifyMessage struct {
	PayStatus int64  `json:"payStatus"`
	OrderSn   string `json:"orderSn"`
}

//发送微信小程序订阅消息.
type SendWxMiniSubMessage struct {
	Openid     string            `json:"openid"`
	TemplateID string            `json:"templateID"`
	Page       string            `json:"page"` //可选 点击模板卡片后的跳转页面，仅限本小程序内的页面。支持带参数,（示例index?foo=bar）。该字段不填则模板无跳转。
	Data       map[string]string `json:"data"` //key1:val1#color1; key2:val2#color2;
}
