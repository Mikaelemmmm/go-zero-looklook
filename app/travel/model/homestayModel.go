package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HomestayModel = (*customHomestayModel)(nil)

type (
	// HomestayModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHomestayModel.
	HomestayModel interface {
		homestayModel
	}

	customHomestayModel struct {
		*defaultHomestayModel
	}
)

// NewHomestayModel returns a model for the database table.
func NewHomestayModel(conn sqlx.SqlConn, c cache.CacheConf) HomestayModel {
	return &customHomestayModel{
		defaultHomestayModel: newHomestayModel(conn, c),
	}
}
