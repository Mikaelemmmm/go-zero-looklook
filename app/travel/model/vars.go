package model

import (
	"errors"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var ErrNotFound = sqlx.ErrNotFound
var ErrNoRowsUpdate = errors.New("update db no rows change")

// 民宿活动类型

var HomestayActivityPreferredType = "preferredHomestay" //优选民宿
var HomestayActivityGoodBusiType = "goodBusiness"       //最佳房东

// 民宿活动上下架

var HomestayActivityDownStatus int64 = 0 //下架
var HomestayActivityUpStatus int64 = 1   //上架
