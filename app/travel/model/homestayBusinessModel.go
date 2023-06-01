package model

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"looklook/common/globalkey"
)

var _ HomestayBusinessModel = (*customHomestayBusinessModel)(nil)

type (
	// HomestayBusinessModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHomestayBusinessModel.
	HomestayBusinessModel interface {
		homestayBusinessModel
		Trans(ctx context.Context, fn func(context context.Context, session sqlx.Session) error) error
		RowBuilder() squirrel.SelectBuilder
		CountBuilder(field string) squirrel.SelectBuilder
		SumBuilder(field string) squirrel.SelectBuilder
		DeleteSoft(ctx context.Context, session sqlx.Session, data *HomestayBusiness) error
		FindOneByQuery(ctx context.Context, rowBuilder squirrel.SelectBuilder) (*HomestayBusiness, error)
		FindAll(ctx context.Context, rowBuilder squirrel.SelectBuilder, orderBy string) ([]*HomestayBusiness, error)
		FindPageListByPage(ctx context.Context, rowBuilder squirrel.SelectBuilder, page, pageSize int64, orderBy string) ([]*HomestayBusiness, error)
		FindPageListByIdDESC(ctx context.Context, rowBuilder squirrel.SelectBuilder, preMinId, pageSize int64) ([]*HomestayBusiness, error)
		FindPageListByIdASC(ctx context.Context, rowBuilder squirrel.SelectBuilder, preMaxId, pageSize int64) ([]*HomestayBusiness, error)
	}

	customHomestayBusinessModel struct {
		*defaultHomestayBusinessModel
	}
)

// NewHomestayBusinessModel returns a model for the database table.
func NewHomestayBusinessModel(conn sqlx.SqlConn, c cache.CacheConf) HomestayBusinessModel {
	return &customHomestayBusinessModel{
		defaultHomestayBusinessModel: newHomestayBusinessModel(conn, c),
	}
}

func (m *defaultHomestayBusinessModel) FindOneByQuery(ctx context.Context, rowBuilder squirrel.SelectBuilder) (*HomestayBusiness, error) {

	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).ToSql()
	if err != nil {
		return nil, err
	}

	var resp HomestayBusiness
	err = m.QueryRowNoCacheCtx(ctx, &resp, query, values...)
	switch err {
	case nil:
		return &resp, nil
	default:
		return nil, err
	}
}

// export logic
func (m *defaultHomestayBusinessModel) RowBuilder() squirrel.SelectBuilder {
	return squirrel.Select(homestayBusinessRows).From(m.table)
}

// export logic
func (m *defaultHomestayBusinessModel) CountBuilder(field string) squirrel.SelectBuilder {
	return squirrel.Select("COUNT(" + field + ")").From(m.table)
}

// export logic
func (m *defaultHomestayBusinessModel) SumBuilder(field string) squirrel.SelectBuilder {
	return squirrel.Select("IFNULL(SUM(" + field + "),0)").From(m.table)
}
