package model

import (
	"database/sql"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

//统一model 执行接口
type Executable interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

var ErrNotFound = sqlx.ErrNotFound

//支付业务类型
var ThirdPaymentServiceTypeHomestayOrder string = "homestayOrder" //民宿支付

//支付方式
var ThirdPaymentPayModelWechatPay = "WECHAT_PAY" //微信支付

//支付状态
var ThirdPaymentPayTradeStateFAIL int64 = -1   //支付失败
var ThirdPaymentPayTradeStateWait int64 = 0    //待支付
var ThirdPaymentPayTradeStateSuccess int64 = 1 //支付成功
var ThirdPaymentPayTradeStateRefund int64 = 2  //已退款
