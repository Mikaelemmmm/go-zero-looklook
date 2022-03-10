package genModel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var ErrNotFound = sqlx.ErrNotFound

//统一model 执行接口
type ModelBuilder interface {
	func (m *default<no value>Model) ListBuilder() squirrel.SelectBuilder

    func (m *default<no value>Model) CountBuilder(field string) squirrel.SelectBuilder

    func (m *default<no value>Model) SumBuilder(field string) squirrel.SelectBuilder 
}