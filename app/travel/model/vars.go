package model

import (
	"database/sql"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 统一model 执行接口
type Executable interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

var ErrNotFound = sqlx.ErrNotFound

// 民宿活动类型
var HomestayActivityPreferredType = "preferredHomestay" //优选民宿
var HomestayActivityGoodBusiType = "goodBusiness"       //最佳房东

// 民宿活动上下架
var HomestayActivityDownStatus int64 = 0 //下架
var HomestayActivityUpStatus int64 = 1   //上架
