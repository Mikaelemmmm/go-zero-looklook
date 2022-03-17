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
	homestayBusinessFieldNames          = builder.RawFieldNames(&HomestayBusiness{})
	homestayBusinessRows                = strings.Join(homestayBusinessFieldNames, ",")
	homestayBusinessRowsExpectAutoSet   = strings.Join(stringx.Remove(homestayBusinessFieldNames, "`id`", "`create_time`", "`update_time`"), ",")
	homestayBusinessRowsWithPlaceHolder = strings.Join(stringx.Remove(homestayBusinessFieldNames, "`id`", "`create_time`", "`update_time`"), "=?,") + "=?"

	cacheLooklookTravelHomestayBusinessIdPrefix     = "cache:looklookTravel:homestayBusiness:id:"
	cacheLooklookTravelHomestayBusinessUserIdPrefix = "cache:looklookTravel:homestayBusiness:userId:"
)

type (
	HomestayBusinessModel interface {
		//新增数据
		Insert(session sqlx.Session, data *HomestayBusiness) (sql.Result, error)
		//根据主键查询一条数据，走缓存
		FindOne(id int64) (*HomestayBusiness, error)
		//根据唯一索引查询一条数据，走缓存
		FindOneByUserId(userId int64) (*HomestayBusiness, error)
		//删除数据
		Delete(session sqlx.Session, id int64) error
		//软删除数据
		DeleteSoft(session sqlx.Session, data *HomestayBusiness) error
		//更新数据
		Update(session sqlx.Session, data *HomestayBusiness) (sql.Result, error)
		//更新数据，使用乐观锁
		UpdateWithVersion(session sqlx.Session, data *HomestayBusiness) error
		//根据条件查询一条数据，不走缓存
		FindOneByQuery(rowBuilder squirrel.SelectBuilder) (*HomestayBusiness, error)
		//sum某个字段
		FindSum(sumBuilder squirrel.SelectBuilder) (float64, error)
		//根据条件统计条数
		FindCount(countBuilder squirrel.SelectBuilder) (int64, error)
		//查询所有数据不分页
		FindAll(rowBuilder squirrel.SelectBuilder, orderBy string) ([]*HomestayBusiness, error)
		//根据页码分页查询分页数据
		FindPageListByPage(rowBuilder squirrel.SelectBuilder, page, pageSize int64, orderBy string) ([]*HomestayBusiness, error)
		//根据id倒序分页查询分页数据
		FindPageListByIdDESC(rowBuilder squirrel.SelectBuilder, preMinId, pageSize int64) ([]*HomestayBusiness, error)
		//根据id升序分页查询分页数据
		FindPageListByIdASC(rowBuilder squirrel.SelectBuilder, preMaxId, pageSize int64) ([]*HomestayBusiness, error)
		//暴露给logic，开启事务
		Trans(fn func(session sqlx.Session) error) error
		//暴露给logic，查询数据的builder
		RowBuilder() squirrel.SelectBuilder
		//暴露给logic，查询count的builder
		CountBuilder(field string) squirrel.SelectBuilder
		//暴露给logic，查询sum的builder
		SumBuilder(field string) squirrel.SelectBuilder
	}

	defaultHomestayBusinessModel struct {
		sqlc.CachedConn
		table string
	}

	HomestayBusiness struct {
		Id          int64     `db:"id"`
		CreateTime  time.Time `db:"create_time"`
		UpdateTime  time.Time `db:"update_time"`
		DeleteTime  time.Time `db:"delete_time"`
		DelState    int64     `db:"del_state"`
		Title       string    `db:"title"`        // 店铺名称
		UserId      int64     `db:"user_id"`      // 关联的用户id
		Info        string    `db:"info"`         // 店铺介绍
		BossInfo    string    `db:"boss_info"`    // 房东介绍
		LicenseFron string    `db:"license_fron"` // 营业执照正面
		LicenseBack string    `db:"license_back"` // 营业执照背面
		RowState    int64     `db:"row_state"`    // 0:禁止营业 1:正常营业
		Star        float64   `db:"star"`         // 店铺整体评价，冗余
		Tags        string    `db:"tags"`         // 每个店家一个标签，自己编辑
		Cover       string    `db:"cover"`        // 封面图
		HeaderImg   string    `db:"header_img"`   // 店招门头图片
		Version     int64     `db:"version"`      // 版本号
	}
)

func NewHomestayBusinessModel(conn sqlx.SqlConn, c cache.CacheConf) HomestayBusinessModel {
	return &defaultHomestayBusinessModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "`homestay_business`",
	}
}

//新增数据
func (m *defaultHomestayBusinessModel) Insert(session sqlx.Session, data *HomestayBusiness) (sql.Result, error) {

	data.DeleteTime = time.Unix(0, 0)

	looklookTravelHomestayBusinessIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayBusinessIdPrefix, data.Id)
	looklookTravelHomestayBusinessUserIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayBusinessUserIdPrefix, data.UserId)
	return m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, homestayBusinessRowsExpectAutoSet)
		if session != nil {
			return session.Exec(query, data.DeleteTime, data.DelState, data.Title, data.UserId, data.Info, data.BossInfo, data.LicenseFron, data.LicenseBack, data.RowState, data.Star, data.Tags, data.Cover, data.HeaderImg, data.Version)
		}
		return conn.Exec(query, data.DeleteTime, data.DelState, data.Title, data.UserId, data.Info, data.BossInfo, data.LicenseFron, data.LicenseBack, data.RowState, data.Star, data.Tags, data.Cover, data.HeaderImg, data.Version)
	}, looklookTravelHomestayBusinessUserIdKey, looklookTravelHomestayBusinessIdKey)

}

//根据主键查询一条数据，走缓存
func (m *defaultHomestayBusinessModel) FindOne(id int64) (*HomestayBusiness, error) {
	looklookTravelHomestayBusinessIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayBusinessIdPrefix, id)
	var resp HomestayBusiness
	err := m.QueryRow(&resp, looklookTravelHomestayBusinessIdKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? and del_state = ? limit 1", homestayBusinessRows, m.table)
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

//根据唯一索引查询一条数据，走缓存
func (m *defaultHomestayBusinessModel) FindOneByUserId(userId int64) (*HomestayBusiness, error) {
	looklookTravelHomestayBusinessUserIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayBusinessUserIdPrefix, userId)
	var resp HomestayBusiness
	err := m.QueryRowIndex(&resp, looklookTravelHomestayBusinessUserIdKey, m.formatPrimary, func(conn sqlx.SqlConn, v interface{}) (i interface{}, e error) {
		query := fmt.Sprintf("select %s from %s where `user_id` = ? and del_state = ?  limit 1", homestayBusinessRows, m.table)
		if err := conn.QueryRow(&resp, query, userId, globalkey.DelStateNo); err != nil {
			return nil, err
		}
		return resp.Id, nil
	}, m.queryPrimary)
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
func (m *defaultHomestayBusinessModel) Update(session sqlx.Session, data *HomestayBusiness) (sql.Result, error) {
	looklookTravelHomestayBusinessUserIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayBusinessUserIdPrefix, data.UserId)
	looklookTravelHomestayBusinessIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayBusinessIdPrefix, data.Id)
	return m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, homestayBusinessRowsWithPlaceHolder)
		if session != nil {
			return session.Exec(query, data.DeleteTime, data.DelState, data.Title, data.UserId, data.Info, data.BossInfo, data.LicenseFron, data.LicenseBack, data.RowState, data.Star, data.Tags, data.Cover, data.HeaderImg, data.Version, data.Id)
		}
		return conn.Exec(query, data.DeleteTime, data.DelState, data.Title, data.UserId, data.Info, data.BossInfo, data.LicenseFron, data.LicenseBack, data.RowState, data.Star, data.Tags, data.Cover, data.HeaderImg, data.Version, data.Id)
	}, looklookTravelHomestayBusinessIdKey, looklookTravelHomestayBusinessUserIdKey)
}

//乐观锁修改数据 ,推荐使用
func (m *defaultHomestayBusinessModel) UpdateWithVersion(session sqlx.Session, data *HomestayBusiness) error {

	oldVersion := data.Version
	data.Version += 1

	var sqlResult sql.Result
	var err error

	looklookTravelHomestayBusinessUserIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayBusinessUserIdPrefix, data.UserId)
	looklookTravelHomestayBusinessIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayBusinessIdPrefix, data.Id)
	sqlResult, err = m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ? and version = ? ", m.table, homestayBusinessRowsWithPlaceHolder)
		if session != nil {
			return session.Exec(query, data.DeleteTime, data.DelState, data.Title, data.UserId, data.Info, data.BossInfo, data.LicenseFron, data.LicenseBack, data.RowState, data.Star, data.Tags, data.Cover, data.HeaderImg, data.Version, data.Id, oldVersion)
		}
		return conn.Exec(query, data.DeleteTime, data.DelState, data.Title, data.UserId, data.Info, data.BossInfo, data.LicenseFron, data.LicenseBack, data.RowState, data.Star, data.Tags, data.Cover, data.HeaderImg, data.Version, data.Id, oldVersion)
	}, looklookTravelHomestayBusinessIdKey, looklookTravelHomestayBusinessUserIdKey)
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
func (m *defaultHomestayBusinessModel) FindOneByQuery(rowBuilder squirrel.SelectBuilder) (*HomestayBusiness, error) {

	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).ToSql()
	if err != nil {
		return nil, err
	}

	var resp HomestayBusiness
	err = m.QueryRowNoCache(&resp, query, values...)
	switch err {
	case nil:
		return &resp, nil
	default:
		return nil, err
	}
}

//统计某个字段总和
func (m *defaultHomestayBusinessModel) FindSum(sumBuilder squirrel.SelectBuilder) (float64, error) {

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
func (m *defaultHomestayBusinessModel) FindCount(countBuilder squirrel.SelectBuilder) (int64, error) {

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
func (m *defaultHomestayBusinessModel) FindAll(rowBuilder squirrel.SelectBuilder, orderBy string) ([]*HomestayBusiness, error) {

	if orderBy == "" {
		rowBuilder = rowBuilder.OrderBy("id DESC")
	} else {
		rowBuilder = rowBuilder.OrderBy(orderBy)
	}

	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).ToSql()
	if err != nil {
		return nil, err
	}

	var resp []*HomestayBusiness
	err = m.QueryRowsNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

//按照页码分页查询数据
func (m *defaultHomestayBusinessModel) FindPageListByPage(rowBuilder squirrel.SelectBuilder, page, pageSize int64, orderBy string) ([]*HomestayBusiness, error) {

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

	var resp []*HomestayBusiness
	err = m.QueryRowsNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

//按照id倒序分页查询数据，不支持排序
func (m *defaultHomestayBusinessModel) FindPageListByIdDESC(rowBuilder squirrel.SelectBuilder, preMinId, pageSize int64) ([]*HomestayBusiness, error) {

	if preMinId > 0 {
		rowBuilder = rowBuilder.Where(" id < ? ", preMinId)
	}

	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).OrderBy("id DESC").Limit(uint64(pageSize)).ToSql()
	if err != nil {
		return nil, err
	}

	var resp []*HomestayBusiness
	err = m.QueryRowsNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

//按照id升序分页查询数据，不支持排序
func (m *defaultHomestayBusinessModel) FindPageListByIdASC(rowBuilder squirrel.SelectBuilder, preMaxId, pageSize int64) ([]*HomestayBusiness, error) {

	if preMaxId > 0 {
		rowBuilder = rowBuilder.Where(" id > ? ", preMaxId)
	}

	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).OrderBy("id ASC").Limit(uint64(pageSize)).ToSql()
	if err != nil {
		return nil, err
	}

	var resp []*HomestayBusiness
	err = m.QueryRowsNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

//暴露给logic查询数据构建条件使用的builder
func (m *defaultHomestayBusinessModel) RowBuilder() squirrel.SelectBuilder {
	return squirrel.Select(homestayBusinessRows).From(m.table)
}

//暴露给logic查询count构建条件使用的builder
func (m *defaultHomestayBusinessModel) CountBuilder(field string) squirrel.SelectBuilder {
	return squirrel.Select("COUNT(" + field + ")").From(m.table)
}

//暴露给logic查询构建条件使用的builder
func (m *defaultHomestayBusinessModel) SumBuilder(field string) squirrel.SelectBuilder {
	return squirrel.Select("IFNULL(SUM(" + field + "),0)").From(m.table)
}

//删除数据
func (m *defaultHomestayBusinessModel) Delete(session sqlx.Session, id int64) error {
	data, err := m.FindOne(id)
	if err != nil {
		return err
	}

	looklookTravelHomestayBusinessIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayBusinessIdPrefix, id)
	looklookTravelHomestayBusinessUserIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayBusinessUserIdPrefix, data.UserId)
	_, err = m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
		if session != nil {
			return session.Exec(query, id)
		}
		return conn.Exec(query, id)
	}, looklookTravelHomestayBusinessIdKey, looklookTravelHomestayBusinessUserIdKey)
	return err
}

//软删除数据
func (m *defaultHomestayBusinessModel) DeleteSoft(session sqlx.Session, data *HomestayBusiness) error {
	data.DelState = globalkey.DelStateYes
	data.DeleteTime = time.Now()
	if err := m.UpdateWithVersion(session, data); err != nil {
		return errors.Wrapf(xerr.NewErrMsg("删除数据失败"), "HomestayBusinessModel delete err : %+v", err)
	}
	return nil
}

//暴露给logic开启事务
func (m *defaultHomestayBusinessModel) Trans(fn func(session sqlx.Session) error) error {

	err := m.Transact(func(session sqlx.Session) error {
		return fn(session)
	})
	return err

}

//格式化缓存key
func (m *defaultHomestayBusinessModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheLooklookTravelHomestayBusinessIdPrefix, primary)
}

//根据主键去db查询一条数据
func (m *defaultHomestayBusinessModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? and del_state = ? limit 1", homestayBusinessRows, m.table)
	return conn.QueryRow(v, query, primary, globalkey.DelStateNo)
}

//!!!!! 其他自定义方法，从此处开始写,此处上方不要写自定义方法!!!!!
