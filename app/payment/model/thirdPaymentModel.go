package model

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"looklook/common/globalkey"
)

var _ ThirdPaymentModel = (*customThirdPaymentModel)(nil)

type (
	// ThirdPaymentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customThirdPaymentModel.
	ThirdPaymentModel interface {
		thirdPaymentModel
		Trans(ctx context.Context, fn func(context context.Context, session sqlx.Session) error) error
		RowBuilder() squirrel.SelectBuilder
		CountBuilder(field string) squirrel.SelectBuilder
		SumBuilder(field string) squirrel.SelectBuilder
		DeleteSoft(ctx context.Context, session sqlx.Session, data *ThirdPayment) error
		FindOneByQuery(ctx context.Context, rowBuilder squirrel.SelectBuilder) (*ThirdPayment, error)
		FindAll(ctx context.Context, rowBuilder squirrel.SelectBuilder, orderBy string) ([]*ThirdPayment, error)
		FindPageListByPage(ctx context.Context, rowBuilder squirrel.SelectBuilder, page, pageSize int64, orderBy string) ([]*ThirdPayment, error)
		FindPageListByIdDESC(ctx context.Context, rowBuilder squirrel.SelectBuilder, preMinId, pageSize int64) ([]*ThirdPayment, error)
		FindPageListByIdASC(ctx context.Context, rowBuilder squirrel.SelectBuilder, preMaxId, pageSize int64) ([]*ThirdPayment, error)
	}

	customThirdPaymentModel struct {
		*defaultThirdPaymentModel
	}
)

// NewThirdPaymentModel returns a model for the database table.
func NewThirdPaymentModel(conn sqlx.SqlConn, c cache.CacheConf) ThirdPaymentModel {
	return &customThirdPaymentModel{
		defaultThirdPaymentModel: newThirdPaymentModel(conn, c),
	}
}

func (m *defaultThirdPaymentModel) FindOneByQuery(ctx context.Context, rowBuilder squirrel.SelectBuilder) (*ThirdPayment, error) {

	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).ToSql()
	if err != nil {
		return nil, err
	}

	var resp ThirdPayment
	err = m.QueryRowNoCacheCtx(ctx, &resp, query, values...)
	switch err {
	case nil:
		return &resp, nil
	default:
		return nil, err
	}
}

// export logic
func (m *defaultThirdPaymentModel) RowBuilder() squirrel.SelectBuilder {
	return squirrel.Select(thirdPaymentRows).From(m.table)
}

// export logic
func (m *defaultThirdPaymentModel) CountBuilder(field string) squirrel.SelectBuilder {
	return squirrel.Select("COUNT(" + field + ")").From(m.table)
}

// export logic
func (m *defaultThirdPaymentModel) SumBuilder(field string) squirrel.SelectBuilder {
	return squirrel.Select("IFNULL(SUM(" + field + "),0)").From(m.table)
}
