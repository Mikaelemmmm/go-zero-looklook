package {{.pkg}}


import (
    "database/sql"
    "github.com/zeromicro/go-zero/core/stores/sqlx"
)

//统一model 执行接口
type Executable interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

var ErrNotFound = sqlx.ErrNotFound
