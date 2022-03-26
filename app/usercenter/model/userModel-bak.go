package model
//
//import (
//	"database/sql"
//	"fmt"
//	"strings"
//	"time"
//
//	"looklook/common/globalkey"
//	"looklook/common/xerr"
//
//	"github.com/Masterminds/squirrel"
//	"github.com/pkg/errors"
//	"github.com/zeromicro/go-zero/core/stores/builder"
//	"github.com/zeromicro/go-zero/core/stores/cache"
//	"github.com/zeromicro/go-zero/core/stores/sqlc"
//	"github.com/zeromicro/go-zero/core/stores/sqlx"
//	"github.com/zeromicro/go-zero/core/stringx"
//)
//
//var (
//	userFieldNames          = builder.RawFieldNames(&User{})
//	userRows                = strings.Join(userFieldNames, ",")
//	userRowsExpectAutoSet   = strings.Join(stringx.Remove(userFieldNames, "`id`", "`create_time`", "`update_time`"), ",")
//	userRowsWithPlaceHolder = strings.Join(stringx.Remove(userFieldNames, "`id`", "`create_time`", "`update_time`"), "=?,") + "=?"
//
//	cacheLooklookUsercenterUserIdPrefix     = "cache:looklookUsercenter:user:id:"
//	cacheLooklookUsercenterUserMobilePrefix = "cache:looklookUsercenter:user:mobile:"
//)
//
//type (
//	UserModel interface {
//		//新增数据
//		Insert(session sqlx.Session, data *User) (sql.Result, error)
//		//根据主键查询一条数据，走缓存
//		FindOne(id int64) (*User, error)
//		//根据唯一索引查询一条数据，走缓存
//		FindOneByMobile(mobile string) (*User, error)
//		//删除数据
//		Delete(session sqlx.Session, id int64) error
//		//软删除数据
//		DeleteSoft(session sqlx.Session, data *User) error
//		//更新数据
//		Update(session sqlx.Session, data *User) (sql.Result, error)
//		//更新数据，使用乐观锁
//		UpdateWithVersion(session sqlx.Session, data *User) error
//		//根据条件查询一条数据，不走缓存
//		FindOneByQuery(rowBuilder squirrel.SelectBuilder) (*User, error)
//		//sum某个字段
//		FindSum(sumBuilder squirrel.SelectBuilder) (float64, error)
//		//根据条件统计条数
//		FindCount(countBuilder squirrel.SelectBuilder) (int64, error)
//		//查询所有数据不分页
//		FindAll(rowBuilder squirrel.SelectBuilder, orderBy string) ([]*User, error)
//		//根据页码分页查询分页数据
//		FindPageListByPage(rowBuilder squirrel.SelectBuilder, page, pageSize int64, orderBy string) ([]*User, error)
//		//根据id倒序分页查询分页数据
//		FindPageListByIdDESC(rowBuilder squirrel.SelectBuilder, preMinId, pageSize int64) ([]*User, error)
//		//根据id升序分页查询分页数据
//		FindPageListByIdASC(rowBuilder squirrel.SelectBuilder, preMaxId, pageSize int64) ([]*User, error)
//		//暴露给logic，开启事务
//		Trans(fn func(session sqlx.Session) error) error
//		//暴露给logic，查询数据的builder
//		RowBuilder() squirrel.SelectBuilder
//		//暴露给logic，查询count的builder
//		CountBuilder(field string) squirrel.SelectBuilder
//		//暴露给logic，查询sum的builder
//		SumBuilder(field string) squirrel.SelectBuilder
//	}
//
//	defaultUserModel struct {
//		sqlc.CachedConn
//		table string
//	}
//
//	User struct {
//		Id         int64     `db:"id"`
//		CreateTime time.Time `db:"create_time"`
//		UpdateTime time.Time `db:"update_time"`
//		DeleteTime time.Time `db:"delete_time"`
//		DelState   int64     `db:"del_state"`
//		Version    int64     `db:"version"` // 版本号
//		Mobile     string    `db:"mobile"`
//		Password   string    `db:"password"`
//		Nickname   string    `db:"nickname"`
//		Sex        int64     `db:"sex"` // 性别 0:男 1:女
//		Avatar     string    `db:"avatar"`
//		Info       string    `db:"info"`
//	}
//)
//
//func NewUserModel(conn sqlx.SqlConn, c cache.CacheConf) UserModel {
//	return &defaultUserModel{
//		CachedConn: sqlc.NewConn(conn, c),
//		table:      "`user`",
//	}
//}
//
////新增数据
//func (m *defaultUserModel) Insert(session sqlx.Session, data *User) (sql.Result, error) {
//
//	data.DeleteTime = time.Unix(0, 0)
//
//	looklookUsercenterUserMobileKey := fmt.Sprintf("%s%v", cacheLooklookUsercenterUserMobilePrefix, data.Mobile)
//	looklookUsercenterUserIdKey := fmt.Sprintf("%s%v", cacheLooklookUsercenterUserIdPrefix, data.Id)
//	return m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
//		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, userRowsExpectAutoSet)
//		if session != nil {
//			return session.Exec(query, data.DeleteTime, data.DelState, data.Version, data.Mobile, data.Password, data.Nickname, data.Sex, data.Avatar, data.Info)
//		}
//		return conn.Exec(query, data.DeleteTime, data.DelState, data.Version, data.Mobile, data.Password, data.Nickname, data.Sex, data.Avatar, data.Info)
//	}, looklookUsercenterUserIdKey, looklookUsercenterUserMobileKey)
//
//}
//
////根据主键查询一条数据，走缓存
//func (m *defaultUserModel) FindOne(id int64) (*User, error) {
//	looklookUsercenterUserIdKey := fmt.Sprintf("%s%v", cacheLooklookUsercenterUserIdPrefix, id)
//	var resp User
//	err := m.QueryRow(&resp, looklookUsercenterUserIdKey, func(conn sqlx.SqlConn, v interface{}) error {
//		query := fmt.Sprintf("select %s from %s where `id` = ? and del_state = ? limit 1", userRows, m.table)
//		return conn.QueryRow(v, query, id, globalkey.DelStateNo)
//	})
//	switch err {
//	case nil:
//		if resp.DelState == globalkey.DelStateYes {
//			return nil, ErrNotFound
//		}
//		return &resp, nil
//	case sqlc.ErrNotFound:
//		return nil, ErrNotFound
//	default:
//		return nil, err
//	}
//}
//
////根据唯一索引查询一条数据，走缓存
//func (m *defaultUserModel) FindOneByMobile(mobile string) (*User, error) {
//	looklookUsercenterUserMobileKey := fmt.Sprintf("%s%v", cacheLooklookUsercenterUserMobilePrefix, mobile)
//	var resp User
//	err := m.QueryRowIndex(&resp, looklookUsercenterUserMobileKey, m.formatPrimary, func(conn sqlx.SqlConn, v interface{}) (i interface{}, e error) {
//		query := fmt.Sprintf("select %s from %s where `mobile` = ? and del_state = ?  limit 1", userRows, m.table)
//		if err := conn.QueryRow(&resp, query, mobile, globalkey.DelStateNo); err != nil {
//			return nil, err
//		}
//		return resp.Id, nil
//	}, m.queryPrimary)
//	switch err {
//	case nil:
//		if resp.DelState == globalkey.DelStateYes {
//			return nil, ErrNotFound
//		}
//		return &resp, nil
//	case sqlc.ErrNotFound:
//		return nil, ErrNotFound
//	default:
//		return nil, err
//	}
//}
//
////修改数据 ,推荐优先使用乐观锁更新
//func (m *defaultUserModel) Update(session sqlx.Session, data *User) (sql.Result, error) {
//	looklookUsercenterUserIdKey := fmt.Sprintf("%s%v", cacheLooklookUsercenterUserIdPrefix, data.Id)
//	looklookUsercenterUserMobileKey := fmt.Sprintf("%s%v", cacheLooklookUsercenterUserMobilePrefix, data.Mobile)
//	return m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
//		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, userRowsWithPlaceHolder)
//		if session != nil {
//			return session.Exec(query, data.DeleteTime, data.DelState, data.Version, data.Mobile, data.Password, data.Nickname, data.Sex, data.Avatar, data.Info, data.Id)
//		}
//		return conn.Exec(query, data.DeleteTime, data.DelState, data.Version, data.Mobile, data.Password, data.Nickname, data.Sex, data.Avatar, data.Info, data.Id)
//	}, looklookUsercenterUserIdKey, looklookUsercenterUserMobileKey)
//}
//
////乐观锁修改数据 ,推荐使用
//func (m *defaultUserModel) UpdateWithVersion(session sqlx.Session, data *User) error {
//
//	oldVersion := data.Version
//	data.Version += 1
//
//	var sqlResult sql.Result
//	var err error
//
//	looklookUsercenterUserIdKey := fmt.Sprintf("%s%v", cacheLooklookUsercenterUserIdPrefix, data.Id)
//	looklookUsercenterUserMobileKey := fmt.Sprintf("%s%v", cacheLooklookUsercenterUserMobilePrefix, data.Mobile)
//	sqlResult, err = m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
//		query := fmt.Sprintf("update %s set %s where `id` = ? and version = ? ", m.table, userRowsWithPlaceHolder)
//		if session != nil {
//			return session.Exec(query, data.DeleteTime, data.DelState, data.Version, data.Mobile, data.Password, data.Nickname, data.Sex, data.Avatar, data.Info, data.Id, oldVersion)
//		}
//		return conn.Exec(query, data.DeleteTime, data.DelState, data.Version, data.Mobile, data.Password, data.Nickname, data.Sex, data.Avatar, data.Info, data.Id, oldVersion)
//	}, looklookUsercenterUserIdKey, looklookUsercenterUserMobileKey)
//	if err != nil {
//		return err
//	}
//
//	updateCount, err := sqlResult.RowsAffected()
//	if err != nil {
//		return err
//	}
//
//	if updateCount == 0 {
//		return xerr.NewErrCode(xerr.DB_UPDATE_AFFECTED_ZERO_ERROR)
//	}
//
//	return nil
//
//}
//
////根据条件查询一条数据
//func (m *defaultUserModel) FindOneByQuery(rowBuilder squirrel.SelectBuilder) (*User, error) {
//
//	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).ToSql()
//	if err != nil {
//		return nil, err
//	}
//
//	var resp User
//	err = m.QueryRowNoCache(&resp, query, values...)
//	switch err {
//	case nil:
//		return &resp, nil
//	default:
//		return nil, err
//	}
//}
//
////统计某个字段总和
//func (m *defaultUserModel) FindSum(sumBuilder squirrel.SelectBuilder) (float64, error) {
//
//	query, values, err := sumBuilder.Where("del_state = ?", globalkey.DelStateNo).ToSql()
//	if err != nil {
//		return 0, err
//	}
//
//	var resp float64
//	err = m.QueryRowNoCache(&resp, query, values...)
//	switch err {
//	case nil:
//		return resp, nil
//	default:
//		return 0, err
//	}
//}
//
////根据某个字段查询数据数量
//func (m *defaultUserModel) FindCount(countBuilder squirrel.SelectBuilder) (int64, error) {
//
//	query, values, err := countBuilder.Where("del_state = ?", globalkey.DelStateNo).ToSql()
//	if err != nil {
//		return 0, err
//	}
//
//	var resp int64
//	err = m.QueryRowNoCache(&resp, query, values...)
//	switch err {
//	case nil:
//		return resp, nil
//	default:
//		return 0, err
//	}
//}
//
////查询所有数据
//func (m *defaultUserModel) FindAll(rowBuilder squirrel.SelectBuilder, orderBy string) ([]*User, error) {
//
//	if orderBy == "" {
//		rowBuilder = rowBuilder.OrderBy("id DESC")
//	} else {
//		rowBuilder = rowBuilder.OrderBy(orderBy)
//	}
//
//	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).ToSql()
//	if err != nil {
//		return nil, err
//	}
//
//	var resp []*User
//	err = m.QueryRowsNoCache(&resp, query, values...)
//	switch err {
//	case nil:
//		return resp, nil
//	default:
//		return nil, err
//	}
//}
//
////按照页码分页查询数据
//func (m *defaultUserModel) FindPageListByPage(rowBuilder squirrel.SelectBuilder, page, pageSize int64, orderBy string) ([]*User, error) {
//
//	if orderBy == "" {
//		rowBuilder = rowBuilder.OrderBy("id DESC")
//	} else {
//		rowBuilder = rowBuilder.OrderBy(orderBy)
//	}
//
//	if page < 1 {
//		page = 1
//	}
//	offset := (page - 1) * pageSize
//
//	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).Offset(uint64(offset)).Limit(uint64(pageSize)).ToSql()
//	if err != nil {
//		return nil, err
//	}
//
//	var resp []*User
//	err = m.QueryRowsNoCache(&resp, query, values...)
//	switch err {
//	case nil:
//		return resp, nil
//	default:
//		return nil, err
//	}
//}
//
////按照id倒序分页查询数据，不支持排序
//func (m *defaultUserModel) FindPageListByIdDESC(rowBuilder squirrel.SelectBuilder, preMinId, pageSize int64) ([]*User, error) {
//
//	if preMinId > 0 {
//		rowBuilder = rowBuilder.Where(" id < ? ", preMinId)
//	}
//
//	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).OrderBy("id DESC").Limit(uint64(pageSize)).ToSql()
//	if err != nil {
//		return nil, err
//	}
//
//	var resp []*User
//	err = m.QueryRowsNoCache(&resp, query, values...)
//	switch err {
//	case nil:
//		return resp, nil
//	default:
//		return nil, err
//	}
//}
//
////按照id升序分页查询数据，不支持排序
//func (m *defaultUserModel) FindPageListByIdASC(rowBuilder squirrel.SelectBuilder, preMaxId, pageSize int64) ([]*User, error) {
//
//	if preMaxId > 0 {
//		rowBuilder = rowBuilder.Where(" id > ? ", preMaxId)
//	}
//
//	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).OrderBy("id ASC").Limit(uint64(pageSize)).ToSql()
//	if err != nil {
//		return nil, err
//	}
//
//	var resp []*User
//	err = m.QueryRowsNoCache(&resp, query, values...)
//	switch err {
//	case nil:
//		return resp, nil
//	default:
//		return nil, err
//	}
//}
//
////暴露给logic查询数据构建条件使用的builder
//func (m *defaultUserModel) RowBuilder() squirrel.SelectBuilder {
//	return squirrel.Select(userRows).From(m.table)
//}
//
////暴露给logic查询count构建条件使用的builder
//func (m *defaultUserModel) CountBuilder(field string) squirrel.SelectBuilder {
//	return squirrel.Select("COUNT(" + field + ")").From(m.table)
//}
//
////暴露给logic查询构建条件使用的builder
//func (m *defaultUserModel) SumBuilder(field string) squirrel.SelectBuilder {
//	return squirrel.Select("IFNULL(SUM(" + field + "),0)").From(m.table)
//}
//
////删除数据
//func (m *defaultUserModel) Delete(session sqlx.Session, id int64) error {
//	data, err := m.FindOne(id)
//	if err != nil {
//		return err
//	}
//
//	looklookUsercenterUserIdKey := fmt.Sprintf("%s%v", cacheLooklookUsercenterUserIdPrefix, id)
//	looklookUsercenterUserMobileKey := fmt.Sprintf("%s%v", cacheLooklookUsercenterUserMobilePrefix, data.Mobile)
//	_, err = m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
//		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
//		if session != nil {
//			return session.Exec(query, id)
//		}
//		return conn.Exec(query, id)
//	}, looklookUsercenterUserIdKey, looklookUsercenterUserMobileKey)
//	return err
//}
//
////软删除数据
//func (m *defaultUserModel) DeleteSoft(session sqlx.Session, data *User) error {
//	data.DelState = globalkey.DelStateYes
//	data.DeleteTime = time.Now()
//	if err := m.UpdateWithVersion(session, data); err != nil {
//		return errors.Wrapf(xerr.NewErrMsg("删除数据失败"), "UserModel delete err : %+v", err)
//	}
//	return nil
//}
//
////暴露给logic开启事务
//func (m *defaultUserModel) Trans(fn func(session sqlx.Session) error) error {
//
//	err := m.Transact(func(session sqlx.Session) error {
//		return fn(session)
//	})
//	return err
//
//}
//
////格式化缓存key
//func (m *defaultUserModel) formatPrimary(primary interface{}) string {
//	return fmt.Sprintf("%s%v", cacheLooklookUsercenterUserIdPrefix, primary)
//}
//
////根据主键去db查询一条数据
//func (m *defaultUserModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
//	query := fmt.Sprintf("select %s from %s where `id` = ? and del_state = ? limit 1", userRows, m.table)
//	return conn.QueryRow(v, query, primary, globalkey.DelStateNo)
//}
//
////!!!!! 其他自定义方法，从此处开始写,此处上方不要写自定义方法!!!!!
