package model

import (
	"database/sql"
	"fmt"
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
	homestayCommentFieldNames          = builder.RawFieldNames(&HomestayComment{})
	homestayCommentRows                = strings.Join(homestayCommentFieldNames, ",")
	homestayCommentRowsExpectAutoSet   = strings.Join(stringx.Remove(homestayCommentFieldNames, "`id`", "`create_time`", "`update_time`"), ",")
	homestayCommentRowsWithPlaceHolder = strings.Join(stringx.Remove(homestayCommentFieldNames, "`id`", "`create_time`", "`update_time`"), "=?,") + "=?"

	cacheLooklookTravelHomestayCommentIdPrefix = "cache:looklookTravel:homestayComment:id:"
)

type (
	HomestayCommentModel interface {
		//新增数据
		Insert(session sqlx.Session, data *HomestayComment) (sql.Result, error)
		//根据主键查询一条数据，走缓存
		FindOne(id int64) (*HomestayComment, error)
		//删除数据
		Delete(session sqlx.Session, id int64) error
		//软删除数据
		DeleteSoft(session sqlx.Session, data *HomestayComment) error
		//更新数据
		Update(session sqlx.Session, data *HomestayComment) (sql.Result, error)
		//更新数据，使用乐观锁
		UpdateWithVersion(session sqlx.Session, data *HomestayComment) error
		//根据条件查询一条数据，不走缓存
		FindOneByQuery(rowBuilder squirrel.SelectBuilder) (*HomestayComment, error)
		//sum某个字段
		FindSum(sumBuilder squirrel.SelectBuilder) (float64, error)
		//根据条件统计条数
		FindCount(countBuilder squirrel.SelectBuilder) (int64, error)
		//查询所有数据不分页
		FindAll(rowBuilder squirrel.SelectBuilder, orderBy string) ([]*HomestayComment, error)
		//根据页码分页查询分页数据
		FindPageListByPage(rowBuilder squirrel.SelectBuilder, page, pageSize int64, orderBy string) ([]*HomestayComment, error)
		//根据id倒序分页查询分页数据
		FindPageListByIdDESC(rowBuilder squirrel.SelectBuilder, preMinId, pageSize int64) ([]*HomestayComment, error)
		//根据id升序分页查询分页数据
		FindPageListByIdASC(rowBuilder squirrel.SelectBuilder, preMaxId, pageSize int64) ([]*HomestayComment, error)
		//暴露给logic，开启事务
		Trans(fn func(session sqlx.Session) error) error
		//暴露给logic，查询数据的builder
		RowBuilder() squirrel.SelectBuilder
		//暴露给logic，查询count的builder
		CountBuilder(field string) squirrel.SelectBuilder
		//暴露给logic，查询sum的builder
		SumBuilder(field string) squirrel.SelectBuilder
	}

	defaultHomestayCommentModel struct {
		sqlc.CachedConn
		table string
	}

	HomestayComment struct {
		Id         int64     `db:"id"`
		CreateTime time.Time `db:"create_time"`
		UpdateTime time.Time `db:"update_time"`
		DeleteTime time.Time `db:"delete_time"`
		DelState   int64     `db:"del_state"`
		HomestayId int64     `db:"homestay_id"` // 民宿id
		UserId     int64     `db:"user_id"`     // 用户id
		Content    string    `db:"content"`     // 评论内容
		Star       string    `db:"star"`        // 星星数,多个维度
		Version    int64     `db:"version"`     // 版本号
	}
)

func NewHomestayCommentModel(conn sqlx.SqlConn, c cache.CacheConf) HomestayCommentModel {
	return &defaultHomestayCommentModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "`homestay_comment`",
	}
}

//新增数据
func (m *defaultHomestayCommentModel) Insert(session sqlx.Session, data *HomestayComment) (sql.Result, error) {

	data.DeleteTime = time.Unix(0, 0)

	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?)", m.table, homestayCommentRowsExpectAutoSet)
	if session != nil {
		return session.Exec(query, data.DeleteTime, data.DelState, data.HomestayId, data.UserId, data.Content, data.Star, data.Version)
	}
	return m.ExecNoCache(query, data.DeleteTime, data.DelState, data.HomestayId, data.UserId, data.Content, data.Star, data.Version)

}

//根据主键查询一条数据，走缓存
func (m *defaultHomestayCommentModel) FindOne(id int64) (*HomestayComment, error) {
	looklookTravelHomestayCommentIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayCommentIdPrefix, id)
	var resp HomestayComment
	err := m.QueryRow(&resp, looklookTravelHomestayCommentIdKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? and del_state = ? limit 1", homestayCommentRows, m.table)
		return conn.QueryRow(v, query, id, globalkey.DelStateNo)
	})
	switch err {
	case nil:
		if resp.DelState == globalkey.DelStateYes {
			return nil, ErrNotFound
		}
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

//修改数据 ,推荐优先使用乐观锁更新
func (m *defaultHomestayCommentModel) Update(session sqlx.Session, data *HomestayComment) (sql.Result, error) {
	looklookTravelHomestayCommentIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayCommentIdPrefix, data.Id)
	return m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, homestayCommentRowsWithPlaceHolder)
		if session != nil {
			return session.Exec(query, data.DeleteTime, data.DelState, data.HomestayId, data.UserId, data.Content, data.Star, data.Version, data.Id)
		}
		return conn.Exec(query, data.DeleteTime, data.DelState, data.HomestayId, data.UserId, data.Content, data.Star, data.Version, data.Id)
	}, looklookTravelHomestayCommentIdKey)
}

//乐观锁修改数据 ,推荐使用
func (m *defaultHomestayCommentModel) UpdateWithVersion(session sqlx.Session, data *HomestayComment) error {

	oldVersion := data.Version
	data.Version += 1

	var sqlResult sql.Result
	var err error

	looklookTravelHomestayCommentIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayCommentIdPrefix, data.Id)
	sqlResult, err = m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ? and version = ? ", m.table, homestayCommentRowsWithPlaceHolder)
		if session != nil {
			return session.Exec(query, data.DeleteTime, data.DelState, data.HomestayId, data.UserId, data.Content, data.Star, data.Version, data.Id, oldVersion)
		}
		return conn.Exec(query, data.DeleteTime, data.DelState, data.HomestayId, data.UserId, data.Content, data.Star, data.Version, data.Id, oldVersion)
	}, looklookTravelHomestayCommentIdKey)
	if err != nil {
		return err
	}

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
func (m *defaultHomestayCommentModel) FindOneByQuery(rowBuilder squirrel.SelectBuilder) (*HomestayComment, error) {

	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).ToSql()
	if err != nil {
		return nil, err
	}

	var resp HomestayComment
	err = m.QueryRowNoCache(&resp, query, values...)
	switch err {
	case nil:
		return &resp, nil
	default:
		return nil, err
	}
}

//统计某个字段总和
func (m *defaultHomestayCommentModel) FindSum(sumBuilder squirrel.SelectBuilder) (float64, error) {

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
func (m *defaultHomestayCommentModel) FindCount(countBuilder squirrel.SelectBuilder) (int64, error) {

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
func (m *defaultHomestayCommentModel) FindAll(rowBuilder squirrel.SelectBuilder, orderBy string) ([]*HomestayComment, error) {

	if orderBy == "" {
		rowBuilder = rowBuilder.OrderBy("id DESC")
	} else {
		rowBuilder = rowBuilder.OrderBy(orderBy)
	}

	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).ToSql()
	if err != nil {
		return nil, err
	}

	var resp []*HomestayComment
	err = m.QueryRowsNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

//按照页码分页查询数据
func (m *defaultHomestayCommentModel) FindPageListByPage(rowBuilder squirrel.SelectBuilder, page, pageSize int64, orderBy string) ([]*HomestayComment, error) {

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

	var resp []*HomestayComment
	err = m.QueryRowsNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

//按照id倒序分页查询数据，不支持排序
func (m *defaultHomestayCommentModel) FindPageListByIdDESC(rowBuilder squirrel.SelectBuilder, preMinId, pageSize int64) ([]*HomestayComment, error) {

	if preMinId > 0 {
		rowBuilder = rowBuilder.Where(" id < ? ", preMinId)
	}

	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).OrderBy("id DESC").Limit(uint64(pageSize)).ToSql()
	if err != nil {
		return nil, err
	}

	var resp []*HomestayComment
	err = m.QueryRowsNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

//按照id升序分页查询数据，不支持排序
func (m *defaultHomestayCommentModel) FindPageListByIdASC(rowBuilder squirrel.SelectBuilder, preMaxId, pageSize int64) ([]*HomestayComment, error) {

	if preMaxId > 0 {
		rowBuilder = rowBuilder.Where(" id > ? ", preMaxId)
	}

	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).OrderBy("id ASC").Limit(uint64(pageSize)).ToSql()
	if err != nil {
		return nil, err
	}

	var resp []*HomestayComment
	err = m.QueryRowsNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

//暴露给logic查询数据构建条件使用的builder
func (m *defaultHomestayCommentModel) RowBuilder() squirrel.SelectBuilder {
	return squirrel.Select(homestayCommentRows).From(m.table)
}

//暴露给logic查询count构建条件使用的builder
func (m *defaultHomestayCommentModel) CountBuilder(field string) squirrel.SelectBuilder {
	return squirrel.Select("COUNT(" + field + ")").From(m.table)
}

//暴露给logic查询构建条件使用的builder
func (m *defaultHomestayCommentModel) SumBuilder(field string) squirrel.SelectBuilder {
	return squirrel.Select("IFNULL(SUM(" + field + "),0)").From(m.table)
}

//删除数据
func (m *defaultHomestayCommentModel) Delete(session sqlx.Session, id int64) error {

	looklookTravelHomestayCommentIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayCommentIdPrefix, id)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
		if session != nil {
			return session.Exec(query, id)
		}
		return conn.Exec(query, id)
	}, looklookTravelHomestayCommentIdKey)
	return err
}

//软删除数据
func (m *defaultHomestayCommentModel) DeleteSoft(session sqlx.Session, data *HomestayComment) error {
	data.DelState = globalkey.DelStateYes
	data.DeleteTime = time.Now()
	if err := m.UpdateWithVersion(session, data); err != nil {
		return errors.Wrapf(xerr.NewErrMsg("删除数据失败"), "HomestayCommentModel delete err : %+v", err)
	}
	return nil
}

//暴露给logic开启事务
func (m *defaultHomestayCommentModel) Trans(fn func(session sqlx.Session) error) error {

	err := m.Transact(func(session sqlx.Session) error {
		return fn(session)
	})
	return err

}

//格式化缓存key
func (m *defaultHomestayCommentModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheLooklookTravelHomestayCommentIdPrefix, primary)
}

//根据主键去db查询一条数据
func (m *defaultHomestayCommentModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? and del_state = ? limit 1", homestayCommentRows, m.table)
	return conn.QueryRow(v, query, primary, globalkey.DelStateNo)
}

//!!!!! 其他自定义方法，从此处开始写,此处上方不要写自定义方法!!!!!
