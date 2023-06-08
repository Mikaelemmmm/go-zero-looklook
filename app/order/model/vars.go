package model

import (
	"errors"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var ErrNotFound = sqlx.ErrNotFound
var ErrNoRowsUpdate = errors.New("update db no rows change")

// HomestayOrder 交易状态 :  -1: 已取消 0:待支付 1:未使用 2:已使用  3:已过期

var HomestayOrderTradeStateCancel int64 = -1
var HomestayOrderTradeStateWaitPay int64 = 0
var HomestayOrderTradeStateWaitUse int64 = 1
var HomestayOrderTradeStateUsed int64 = 2
var HomestayOrderTradeStateRefund int64 = 3
var HomestayOrderTradeStateExpire int64 = 4

// 是否需要餐食

var HomestayOrderNeedFoodNo int64 = 0
var HomestayOrderNeedFoodYes int64 = 1
