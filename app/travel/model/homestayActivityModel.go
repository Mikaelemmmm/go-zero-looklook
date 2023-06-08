package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HomestayActivityModel = (*customHomestayActivityModel)(nil)

type (
	// HomestayActivityModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHomestayActivityModel.
	HomestayActivityModel interface {
		homestayActivityModel
	}

	customHomestayActivityModel struct {
		*defaultHomestayActivityModel
	}
)

// NewHomestayActivityModel returns a model for the database table.
func NewHomestayActivityModel(conn sqlx.SqlConn, c cache.CacheConf) HomestayActivityModel {
	return &customHomestayActivityModel{
		defaultHomestayActivityModel: newHomestayActivityModel(conn, c),
	}
}
