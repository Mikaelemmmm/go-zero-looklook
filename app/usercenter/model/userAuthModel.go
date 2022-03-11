package model

import (
	"database/sql"
	"fmt"
	"looklook/deploy/script/mysql/genModel"
	"strings"
	"time"

	"looklook/common/globalkey"
	"looklook/common/xerr"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	userAuthFieldNames          = builder.RawFieldNames(&UserAuth{})
	userAuthRows                = strings.Join(userAuthFieldNames, ",")
	userAuthRowsExpectAutoSet   = strings.Join(stringx.Remove(userAuthFieldNames, "`id`", "`create_time`", "`update_time`"), ",")
	userAuthRowsWithPlaceHolder = strings.Join(stringx.Remove(userAuthFieldNames, "`id`", "`create_time`", "`update_time`"), "=?,") + "=?"

	cacheLooklookUsercenterUserAuthIdPrefix              = "cache:looklookUsercenter:userAuth:id:"
	cacheLooklookUsercenterUserAuthAuthTypeAuthKeyPrefix = "cache:looklookUsercenter:userAuth:authType:authKey:"
	cacheLooklookUsercenterUserAuthUserIdAuthTypePrefix  = "cache:looklookUsercenter:userAuth:userId:authType:"
)

type (
	UserAuthModel interface {
		//根据主键查询一条数据，走缓存
		FindOne(id int64) (*UserAuth, error)
		//根据唯一索引查询一条数据，走缓存
		FindOneByAuthTypeAuthKey(authType string, authKey string) (*UserAuth, error)
		//根据唯一索引查询一条数据，走缓存
		FindOneByUserIdAuthType(userId int64, authType string) (*UserAuth, error)
		//新增数据
		Insert(session sqlx.Session, data *UserAuth) (sql.Result, error)
		//删除数据
		Delete(session sqlx.Session, data *UserAuth) error
		//更新数据
		Update(session sqlx.Session, data *UserAuth) (sql.Result, error)
		//更新数据，使用乐观锁
		UpdateWithVersion(session sqlx.Session, data *UserAuth) error
		//根据条件查询一条数据，不走缓存
		FindOneByQuery(rowBuilder squirrel.SelectBuilder) (*UserAuth, error)
		//sum某个字段
		FindSum(sumBuilder squirrel.SelectBuilder) (float64, error)
		//根据条件统计条数
		FindCount(countBuilder squirrel.SelectBuilder) (int64, error)
		//查询所有数据不分页
		FindAll(rowBuilder squirrel.SelectBuilder, orderBy string) ([]*UserAuth, error)
		//根据页码分页查询分页数据
		FindPageListByPage(rowBuilder squirrel.SelectBuilder, page, pageSize int64, orderBy string) ([]*UserAuth, error)
		//根据id倒序分页查询分页数据
		FindPageListByIdDESC(rowBuilder squirrel.SelectBuilder, preMinId, pageSize int64) ([]*UserAuth, error)
		//根据id升序分页查询分页数据
		FindPageListByIdASC(rowBuilder squirrel.SelectBuilder, preMaxId, pageSize int64) ([]*UserAuth, error)
		//暴露给logic，开启事务
		Trans(fn func(session sqlx.Session) error) error
		//暴露给logic，查询数据的builder
		RowBuilder() squirrel.SelectBuilder
		//暴露给logic，查询count的builder
		CountBuilder(field string) squirrel.SelectBuilder
		//暴露给logic，查询sum的builder
		SumBuilder(field string) squirrel.SelectBuilder
	}

	defaultUserAuthModel struct {
		sqlc.CachedConn
		table string
	}

	UserAuth struct {
		Id         int64     `db:"id"`
		CreateTime time.Time `db:"create_time"`
		UpdateTime time.Time `db:"update_time"`
		DeleteTime time.Time `db:"delete_time"`
		DelState   int64     `db:"del_state"`
		Version    int64     `db:"version"` // 版本号
		UserId     int64     `db:"user_id"`
		AuthKey    string    `db:"auth_key"`  // 平台唯一id
		AuthType   string    `db:"auth_type"` // 平台类型
	}
)

func NewUserAuthModel(conn sqlx.SqlConn, c cache.CacheConf) UserAuthModel {
	return &defaultUserAuthModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "`user_auth`",
	}
}

//新增数据
func (m *defaultUserAuthModel) Insert(session sqlx.Session, data *UserAuth) (sql.Result, error) {

	data.DeleteTime = time.Unix(0, 0)

	looklookUsercenterUserAuthIdKey := fmt.Sprintf("%s%v", cacheLooklookUsercenterUserAuthIdPrefix, data.Id)
	looklookUsercenterUserAuthAuthTypeAuthKeyKey := fmt.Sprintf("%s%v:%v", cacheLooklookUsercenterUserAuthAuthTypeAuthKeyPrefix, data.AuthType, data.AuthKey)
	looklookUsercenterUserAuthUserIdAuthTypeKey := fmt.Sprintf("%s%v:%v", cacheLooklookUsercenterUserAuthUserIdAuthTypePrefix, data.UserId, data.AuthType)
	return m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?)", m.table, userAuthRowsExpectAutoSet)
		if session != nil {
			return session.Exec(query, data.DeleteTime, data.DelState, data.Version, data.UserId, data.AuthKey, data.AuthType)
		}
		return conn.Exec(query, data.DeleteTime, data.DelState, data.Version, data.UserId, data.AuthKey, data.AuthType)
	}, looklookUsercenterUserAuthIdKey, looklookUsercenterUserAuthAuthTypeAuthKeyKey, looklookUsercenterUserAuthUserIdAuthTypeKey)

}

//根据主键查询一条数据，走缓存
func (m *defaultUserAuthModel) FindOne(id int64) (*UserAuth, error) {
	looklookUsercenterUserAuthIdKey := fmt.Sprintf("%s%v", cacheLooklookUsercenterUserAuthIdPrefix, id)
	var resp UserAuth
	err := m.QueryRow(&resp, looklookUsercenterUserAuthIdKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? and del_state = ? limit 1", userAuthRows, m.table)
		return conn.QueryRow(v, query, id, globalkey.DelStateNo)
	})
	switch err {
	case nil:
		if resp.DelState == globalkey.DelStateYes {
			return nil, genModel.ErrNotFound
		}
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, genModel.ErrNotFound
	default:
		return nil, err
	}
}

//根据唯一索引查询一条数据，走缓存
func (m *defaultUserAuthModel) FindOneByAuthTypeAuthKey(authType string, authKey string) (*UserAuth, error) {
	looklookUsercenterUserAuthAuthTypeAuthKeyKey := fmt.Sprintf("%s%v:%v", cacheLooklookUsercenterUserAuthAuthTypeAuthKeyPrefix, authType, authKey)
	var resp UserAuth
	err := m.QueryRowIndex(&resp, looklookUsercenterUserAuthAuthTypeAuthKeyKey, m.formatPrimary, func(conn sqlx.SqlConn, v interface{}) (i interface{}, e error) {
		query := fmt.Sprintf("select %s from %s where `auth_type` = ? and `auth_key` = ? and del_state = ?  limit 1", userAuthRows, m.table)
		if err := conn.QueryRow(&resp, query, authType, authKey, globalkey.DelStateNo); err != nil {
			return nil, err
		}
		return resp.Id, nil
	}, m.queryPrimary)
	switch err {
	case nil:
		if resp.DelState == globalkey.DelStateYes {
			return nil, genModel.ErrNotFound
		}
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, genModel.ErrNotFound
	default:
		return nil, err
	}
}

//根据唯一索引查询一条数据，走缓存
func (m *defaultUserAuthModel) FindOneByUserIdAuthType(userId int64, authType string) (*UserAuth, error) {
	looklookUsercenterUserAuthUserIdAuthTypeKey := fmt.Sprintf("%s%v:%v", cacheLooklookUsercenterUserAuthUserIdAuthTypePrefix, userId, authType)
	var resp UserAuth
	err := m.QueryRowIndex(&resp, looklookUsercenterUserAuthUserIdAuthTypeKey, m.formatPrimary, func(conn sqlx.SqlConn, v interface{}) (i interface{}, e error) {
		query := fmt.Sprintf("select %s from %s where `user_id` = ? and `auth_type` = ? and del_state = ?  limit 1", userAuthRows, m.table)
		if err := conn.QueryRow(&resp, query, userId, authType, globalkey.DelStateNo); err != nil {
			return nil, err
		}
		return resp.Id, nil
	}, m.queryPrimary)
	switch err {
	case nil:
		if resp.DelState == globalkey.DelStateYes {
			return nil, genModel.ErrNotFound
		}
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, genModel.ErrNotFound
	default:
		return nil, err
	}
}

//修改数据 ,推荐优先使用乐观锁更新
func (m *defaultUserAuthModel) Update(session sqlx.Session, data *UserAuth) (sql.Result, error) {
	looklookUsercenterUserAuthUserIdAuthTypeKey := fmt.Sprintf("%s%v:%v", cacheLooklookUsercenterUserAuthUserIdAuthTypePrefix, data.UserId, data.AuthType)
	looklookUsercenterUserAuthIdKey := fmt.Sprintf("%s%v", cacheLooklookUsercenterUserAuthIdPrefix, data.Id)
	looklookUsercenterUserAuthAuthTypeAuthKeyKey := fmt.Sprintf("%s%v:%v", cacheLooklookUsercenterUserAuthAuthTypeAuthKeyPrefix, data.AuthType, data.AuthKey)
	return m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, userAuthRowsWithPlaceHolder)
		if session != nil {
			return session.Exec(query, data.DeleteTime, data.DelState, data.Version, data.UserId, data.AuthKey, data.AuthType, data.Id)
		}
		return conn.Exec(query, data.DeleteTime, data.DelState, data.Version, data.UserId, data.AuthKey, data.AuthType, data.Id)
	}, looklookUsercenterUserAuthIdKey, looklookUsercenterUserAuthAuthTypeAuthKeyKey, looklookUsercenterUserAuthUserIdAuthTypeKey)
}

//乐观锁修改数据 ,推荐使用
func (m *defaultUserAuthModel) UpdateWithVersion(session sqlx.Session, data *UserAuth) error {

	oldVersion := data.Version
	data.Version += 1

	var sqlResult sql.Result
	var err error

	looklookUsercenterUserAuthUserIdAuthTypeKey := fmt.Sprintf("%s%v:%v", cacheLooklookUsercenterUserAuthUserIdAuthTypePrefix, data.UserId, data.AuthType)
	looklookUsercenterUserAuthIdKey := fmt.Sprintf("%s%v", cacheLooklookUsercenterUserAuthIdPrefix, data.Id)
	looklookUsercenterUserAuthAuthTypeAuthKeyKey := fmt.Sprintf("%s%v:%v", cacheLooklookUsercenterUserAuthAuthTypeAuthKeyPrefix, data.AuthType, data.AuthKey)
	sqlResult, err = m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ? and version = ? ", m.table, userAuthRowsWithPlaceHolder)
		if session != nil {
			return session.Exec(query, data.DeleteTime, data.DelState, data.Version, data.UserId, data.AuthKey, data.AuthType, data.Id, oldVersion)
		}
		return conn.Exec(query, data.DeleteTime, data.DelState, data.Version, data.UserId, data.AuthKey, data.AuthType, data.Id, oldVersion)
	}, looklookUsercenterUserAuthIdKey, looklookUsercenterUserAuthAuthTypeAuthKeyKey, looklookUsercenterUserAuthUserIdAuthTypeKey)

	updateCount, err := sqlResult.RowsAffected()
	if err != nil {
		return err
	}

	if updateCount == 0 {
		return xerr.NewErrCode(xerr.DB_UPDATE_AFFECTED_ZERO_ERROR)
	}

	return nil

}

//根据条件查询一条数据
func (m *defaultUserAuthModel) FindOneByQuery(rowBuilder squirrel.SelectBuilder) (*UserAuth, error) {

	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).ToSql()
	if err != nil {
		return nil, err
	}

	var resp UserAuth
	err = m.QueryRowNoCache(&resp, query, values...)
	switch err {
	case nil:
		return &resp, nil
	default:
		return nil, err
	}
}

//统计某个字段总和
func (m *defaultUserAuthModel) FindSum(sumBuilder squirrel.SelectBuilder) (float64, error) {

	query, values, err := sumBuilder.Where("del_state = ?", globalkey.DelStateNo).ToSql()
	if err != nil {
		return 0, err
	}

	var resp float64
	err = m.QueryRowNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return 0, err
	}
}

//根据某个字段查询数据数量
func (m *defaultUserAuthModel) FindCount(countBuilder squirrel.SelectBuilder) (int64, error) {

	query, values, err := countBuilder.Where("del_state = ?", globalkey.DelStateNo).ToSql()
	if err != nil {
		return 0, err
	}

	var resp int64
	err = m.QueryRowNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return 0, err
	}
}

//查询所有数据
func (m *defaultUserAuthModel) FindAll(rowBuilder squirrel.SelectBuilder, orderBy string) ([]*UserAuth, error) {

	if orderBy == "" {
		rowBuilder = rowBuilder.OrderBy("id DESC")
	} else {
		rowBuilder = rowBuilder.OrderBy(orderBy)
	}

	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).ToSql()
	if err != nil {
		return nil, err
	}

	var resp []*UserAuth
	err = m.QueryRowsNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

//按照页码分页查询数据
func (m *defaultUserAuthModel) FindPageListByPage(rowBuilder squirrel.SelectBuilder, page, pageSize int64, orderBy string) ([]*UserAuth, error) {

	if orderBy == "" {
		rowBuilder = rowBuilder.OrderBy("id DESC")
	} else {
		rowBuilder = rowBuilder.OrderBy(orderBy)
	}

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).Offset(uint64(offset)).Limit(uint64(pageSize)).ToSql()
	if err != nil {
		return nil, err
	}

	var resp []*UserAuth
	err = m.QueryRowsNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

//按照id倒序分页查询数据，不支持排序
func (m *defaultUserAuthModel) FindPageListByIdDESC(rowBuilder squirrel.SelectBuilder, preMinId, pageSize int64) ([]*UserAuth, error) {

	if preMinId > 0 {
		rowBuilder = rowBuilder.Where(" id < ? ", preMinId)
	}

	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).OrderBy("id DESC").Limit(uint64(pageSize)).ToSql()
	if err != nil {
		return nil, err
	}

	var resp []*UserAuth
	err = m.QueryRowsNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

//按照id升序分页查询数据，不支持排序
func (m *defaultUserAuthModel) FindPageListByIdASC(rowBuilder squirrel.SelectBuilder, preMaxId, pageSize int64) ([]*UserAuth, error) {

	if preMaxId > 0 {
		rowBuilder = rowBuilder.Where(" id > ? ", preMaxId)
	}

	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).OrderBy("id ASC").Limit(uint64(pageSize)).ToSql()
	if err != nil {
		return nil, err
	}

	var resp []*UserAuth
	err = m.QueryRowsNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

//暴露给logic查询数据构建条件使用的builder
func (m *defaultUserAuthModel) RowBuilder() squirrel.SelectBuilder {
	return squirrel.Select(userAuthRows).From(m.table)
}

//暴露给logic查询count构建条件使用的builder
func (m *defaultUserAuthModel) CountBuilder(field string) squirrel.SelectBuilder {
	return squirrel.Select("COUNT(" + field + ")").From(m.table)
}

//暴露给logic查询构建条件使用的builder
func (m *defaultUserAuthModel) SumBuilder(field string) squirrel.SelectBuilder {
	return squirrel.Select("IFNULL(SUM(" + field + "),0)").From(m.table)
}

//删除数据
func (m *defaultUserAuthModel) Delete(session sqlx.Session, data *UserAuth) error {
	data.DelState = globalkey.DelStateYes
	data.DeleteTime = time.Now()
	if err := m.UpdateWithVersion(session, data); err != nil {
		return errors.Wrapf(xerr.NewErrMsg("删除数据失败"), "UserAuthModel delete err : %+v", err)
	}
	return nil
}

//暴露给logic开启事务
func (m *defaultUserAuthModel) Trans(fn func(session sqlx.Session) error) error {

	err := m.Transact(func(session sqlx.Session) error {
		return fn(session)
	})
	return err

}

//格式化缓存key
func (m *defaultUserAuthModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheLooklookUsercenterUserAuthIdPrefix, primary)
}

//根据主键去db查询一条数据
func (m *defaultUserAuthModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? and del_state = ? limit 1", userAuthRows, m.table)
	return conn.QueryRow(v, query, primary, globalkey.DelStateNo)
}

//!!!!! 其他自定义方法，从此处开始写,此处上方不要写自定义方法!!!!!
