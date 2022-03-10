package genModel

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
	thirdPaymentFieldNames          = builder.RawFieldNames(&ThirdPayment{})
	thirdPaymentRows                = strings.Join(thirdPaymentFieldNames, ",")
	thirdPaymentRowsExpectAutoSet   = strings.Join(stringx.Remove(thirdPaymentFieldNames, "`id`", "`create_time`", "`update_time`"), ",")
	thirdPaymentRowsWithPlaceHolder = strings.Join(stringx.Remove(thirdPaymentFieldNames, "`id`", "`create_time`", "`update_time`"), "=?,") + "=?"

	cacheLooklookPaymentThirdPaymentIdPrefix = "cache:looklookPayment:thirdPayment:id:"
	cacheLooklookPaymentThirdPaymentSnPrefix = "cache:looklookPayment:thirdPayment:sn:"
)

type (
	ThirdPaymentModel interface {
		//根据主键查询一条数据，走缓存
		FindOne(id int64) (*ThirdPayment, error)
		//根据唯一索引查询一条数据，走缓存
		FindOneBySn(sn string) (*ThirdPayment, error)
		//新增数据
		Insert(session sqlx.Session, data *ThirdPayment) (sql.Result, error)
		//删除数据
		Delete(session sqlx.Session, data *ThirdPayment) error
		//更新数据
		Update(session sqlx.Session, data *ThirdPayment) (sql.Result, error)
		//更新数据，使用乐观锁
		UpdateWithVersion(session sqlx.Session, data *ThirdPayment) error
		//根据条件查询一条数据，不走缓存
		FindOneByQuery(sumBuilder squirrel.SelectBuilder) (*ThirdPayment, error)
		//sum某个字段
		FindSum(sumBuilder squirrel.SelectBuilder) (float64, error)
		//根据条件统计条数
		FindCount(countBuilder squirrel.SelectBuilder) (int64, error)
		//查询所有数据不分页
		FindAll(sqlBuilder squirrel.SelectBuilder, orderBy string) ([]*ThirdPayment, error)
		//根据页码分页查询分页数据
		FindPageListByPage(sqlBuilder squirrel.SelectBuilder, page, pageSize int64, orderBy string) ([]*ThirdPayment, error)
		//根据id倒序分页查询分页数据
		FindPageListByIdDESC(sqlBuilder squirrel.SelectBuilder, preMinId, pageSize int64) ([]*ThirdPayment, error)
		//根据id升序分页查询分页数据
		FindPageListByIdASC(sqlBuilder squirrel.SelectBuilder, preMaxId, pageSize int64) ([]*ThirdPayment, error)
		//暴露给logic，开启事务
		Trans(fn func(session sqlx.Session) error) error
		//暴露给logic，查询数据的builder
		RowBuilder() squirrel.SelectBuilder
		//暴露给logic，查询count的builder
		CountBuilder(field string) squirrel.SelectBuilder
		//暴露给logic，查询sum的builder
		SumBuilder(field string) squirrel.SelectBuilder
	}

	defaultThirdPaymentModel struct {
		sqlc.CachedConn
		table string
	}

	ThirdPayment struct {
		Id             int64     `db:"id"`
		Sn             string    `db:"sn"` // 流水单号
		CreateTime     time.Time `db:"create_time"`
		UpdateTime     time.Time `db:"update_time"`
		DeleteTime     time.Time `db:"delete_time"`
		DelState       int64     `db:"del_state"`
		Version        int64     `db:"version"`          // 乐观锁版本号
		UserId         int64     `db:"user_id"`          // 用户id
		PayMode        string    `db:"pay_mode"`         // 支付方式 1:微信支付
		TradeType      string    `db:"trade_type"`       // 第三方支付类型
		TradeState     string    `db:"trade_state"`      // 第三方交易状态
		PayTotal       int64     `db:"pay_total"`        // 支付总金额(分)
		TransactionId  string    `db:"transaction_id"`   // 第三方支付单号
		TradeStateDesc string    `db:"trade_state_desc"` // 支付状态描述
		OrderSn        string    `db:"order_sn"`         // 业务单号
		ServiceType    string    `db:"service_type"`     // 业务类型
		PayStatus      int64     `db:"pay_status"`       // 平台内交易状态   -1:支付失败 0:未支付 1:支付成功 2:已退款
		PayTime        time.Time `db:"pay_time"`         // 支付成功时间
	}
)

func NewThirdPaymentModel(conn sqlx.SqlConn, c cache.CacheConf) ThirdPaymentModel {
	return &defaultThirdPaymentModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "`third_payment`",
	}
}

//新增数据
func (m *defaultThirdPaymentModel) Insert(session sqlx.Session, data *ThirdPayment) (sql.Result, error) {

	data.DeleteTime = time.Unix(0, 0)

	looklookPaymentThirdPaymentIdKey := fmt.Sprintf("%s%v", cacheLooklookPaymentThirdPaymentIdPrefix, data.Id)
	looklookPaymentThirdPaymentSnKey := fmt.Sprintf("%s%v", cacheLooklookPaymentThirdPaymentSnPrefix, data.Sn)
	return m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, thirdPaymentRowsExpectAutoSet)
		if session != nil {
			return session.Exec(query, data.Sn, data.DeleteTime, data.DelState, data.Version, data.UserId, data.PayMode, data.TradeType, data.TradeState, data.PayTotal, data.TransactionId, data.TradeStateDesc, data.OrderSn, data.ServiceType, data.PayStatus, data.PayTime)
		}
		return conn.Exec(query, data.Sn, data.DeleteTime, data.DelState, data.Version, data.UserId, data.PayMode, data.TradeType, data.TradeState, data.PayTotal, data.TransactionId, data.TradeStateDesc, data.OrderSn, data.ServiceType, data.PayStatus, data.PayTime)
	}, looklookPaymentThirdPaymentIdKey, looklookPaymentThirdPaymentSnKey)

}

//根据主键查询一条数据，走缓存
func (m *defaultThirdPaymentModel) FindOne(id int64) (*ThirdPayment, error) {
	looklookPaymentThirdPaymentIdKey := fmt.Sprintf("%s%v", cacheLooklookPaymentThirdPaymentIdPrefix, id)
	var resp ThirdPayment
	err := m.QueryRow(&resp, looklookPaymentThirdPaymentIdKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? and del_state = ? limit 1", thirdPaymentRows, m.table)
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
func (m *defaultThirdPaymentModel) FindOneBySn(sn string) (*ThirdPayment, error) {
	looklookPaymentThirdPaymentSnKey := fmt.Sprintf("%s%v", cacheLooklookPaymentThirdPaymentSnPrefix, sn)
	var resp ThirdPayment
	err := m.QueryRowIndex(&resp, looklookPaymentThirdPaymentSnKey, m.formatPrimary, func(conn sqlx.SqlConn, v interface{}) (i interface{}, e error) {
		query := fmt.Sprintf("select %s from %s where `sn` = ? and del_state = ?  limit 1", thirdPaymentRows, m.table)
		if err := conn.QueryRow(&resp, query, sn, globalkey.DelStateNo); err != nil {
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
func (m *defaultThirdPaymentModel) Update(session sqlx.Session, data *ThirdPayment) (sql.Result, error) {
	looklookPaymentThirdPaymentIdKey := fmt.Sprintf("%s%v", cacheLooklookPaymentThirdPaymentIdPrefix, data.Id)
	looklookPaymentThirdPaymentSnKey := fmt.Sprintf("%s%v", cacheLooklookPaymentThirdPaymentSnPrefix, data.Sn)
	return m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, thirdPaymentRowsWithPlaceHolder)
		if session != nil {
			return session.Exec(query, data.Sn, data.DeleteTime, data.DelState, data.Version, data.UserId, data.PayMode, data.TradeType, data.TradeState, data.PayTotal, data.TransactionId, data.TradeStateDesc, data.OrderSn, data.ServiceType, data.PayStatus, data.PayTime, data.Id)
		}
		return conn.Exec(query, data.Sn, data.DeleteTime, data.DelState, data.Version, data.UserId, data.PayMode, data.TradeType, data.TradeState, data.PayTotal, data.TransactionId, data.TradeStateDesc, data.OrderSn, data.ServiceType, data.PayStatus, data.PayTime, data.Id)
	}, looklookPaymentThirdPaymentIdKey, looklookPaymentThirdPaymentSnKey)
}

//乐观锁修改数据 ,推荐使用
func (m *defaultThirdPaymentModel) UpdateWithVersion(session sqlx.Session, data *ThirdPayment) error {

	oldVersion := data.Version
	data.Version += 1

	var sqlResult sql.Result
	var err error

	looklookPaymentThirdPaymentIdKey := fmt.Sprintf("%s%v", cacheLooklookPaymentThirdPaymentIdPrefix, data.Id)
	looklookPaymentThirdPaymentSnKey := fmt.Sprintf("%s%v", cacheLooklookPaymentThirdPaymentSnPrefix, data.Sn)
	sqlResult, err = m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ? and version = ? ", m.table, thirdPaymentRowsWithPlaceHolder)
		if session != nil {
			return session.Exec(query, data.Sn, data.DeleteTime, data.DelState, data.Version, data.UserId, data.PayMode, data.TradeType, data.TradeState, data.PayTotal, data.TransactionId, data.TradeStateDesc, data.OrderSn, data.ServiceType, data.PayStatus, data.PayTime, data.Id, oldVersion)
		}
		return conn.Exec(query, data.Sn, data.DeleteTime, data.DelState, data.Version, data.UserId, data.PayMode, data.TradeType, data.TradeState, data.PayTotal, data.TransactionId, data.TradeStateDesc, data.OrderSn, data.ServiceType, data.PayStatus, data.PayTime, data.Id, oldVersion)
	}, looklookPaymentThirdPaymentIdKey, looklookPaymentThirdPaymentSnKey)

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
func (m *defaultThirdPaymentModel) FindOneByQuery(sumBuilder squirrel.SelectBuilder) (*ThirdPayment, error) {

	query, values, err := sumBuilder.Where("del_state = ?", globalkey.DelStateNo).ToSql()
	if err != nil {
		return nil, err
	}

	var resp ThirdPayment
	err = m.QueryRowNoCache(&resp, query, values...)
	switch err {
	case nil:
		return &resp, nil
	default:
		return nil, err
	}
}

//统计某个字段总和
func (m *defaultThirdPaymentModel) FindSum(sumBuilder squirrel.SelectBuilder) (float64, error) {

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
func (m *defaultThirdPaymentModel) FindCount(countBuilder squirrel.SelectBuilder) (int64, error) {

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
func (m *defaultThirdPaymentModel) FindAll(sqlBuilder squirrel.SelectBuilder, orderBy string) ([]*ThirdPayment, error) {

	if orderBy == "" {
		sqlBuilder = sqlBuilder.OrderBy("id DESC")
	} else {
		sqlBuilder = sqlBuilder.OrderBy(orderBy)
	}

	query, values, err := sqlBuilder.Where("del_state = ?", globalkey.DelStateNo).ToSql()
	if err != nil {
		return nil, err
	}

	var resp []*ThirdPayment
	err = m.QueryRowsNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

//按照页码分页查询数据
func (m *defaultThirdPaymentModel) FindPageListByPage(sqlBuilder squirrel.SelectBuilder, page, pageSize int64, orderBy string) ([]*ThirdPayment, error) {

	if orderBy == "" {
		sqlBuilder = sqlBuilder.OrderBy("id DESC")
	} else {
		sqlBuilder = sqlBuilder.OrderBy(orderBy)
	}

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	query, values, err := sqlBuilder.Where("del_state = ?", globalkey.DelStateNo).Offset(uint64(offset)).Limit(uint64(pageSize)).ToSql()
	if err != nil {
		return nil, err
	}

	var resp []*ThirdPayment
	err = m.QueryRowsNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

//按照id倒序分页查询数据，不支持排序
func (m *defaultThirdPaymentModel) FindPageListByIdDESC(sqlBuilder squirrel.SelectBuilder, preMinId, pageSize int64) ([]*ThirdPayment, error) {

	if preMinId > 0 {
		sqlBuilder = sqlBuilder.Where(" id < ? ", preMinId)
	}

	query, values, err := sqlBuilder.Where("del_state = ?", globalkey.DelStateNo).OrderBy("id DESC").Limit(uint64(pageSize)).ToSql()
	if err != nil {
		return nil, err
	}

	var resp []*ThirdPayment
	err = m.QueryRowsNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

//按照id升序分页查询数据，不支持排序
func (m *defaultThirdPaymentModel) FindPageListByIdASC(sqlBuilder squirrel.SelectBuilder, preMaxId, pageSize int64) ([]*ThirdPayment, error) {

	if preMaxId > 0 {
		sqlBuilder = sqlBuilder.Where(" id > ? ", preMaxId)
	}

	query, values, err := sqlBuilder.Where("del_state = ?", globalkey.DelStateNo).OrderBy("id ASC").Limit(uint64(pageSize)).ToSql()
	if err != nil {
		return nil, err
	}

	var resp []*ThirdPayment
	err = m.QueryRowsNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

//暴露给logic查询数据构建条件使用的builder
func (m *defaultThirdPaymentModel) RowBuilder() squirrel.SelectBuilder {
	return squirrel.Select(thirdPaymentRows).From(m.table)
}

//暴露给logic查询count构建条件使用的builder
func (m *defaultThirdPaymentModel) CountBuilder(field string) squirrel.SelectBuilder {
	return squirrel.Select("COUNT(" + field + ")").From(m.table)
}

//暴露给logic查询构建条件使用的builder
func (m *defaultThirdPaymentModel) SumBuilder(field string) squirrel.SelectBuilder {
	return squirrel.Select("IFNULL(SUM(" + field + "),0)").From(m.table)
}

//删除数据
func (m *defaultThirdPaymentModel) Delete(session sqlx.Session, data *ThirdPayment) error {
	data.DelState = globalkey.DelStateYes
	data.DeleteTime = time.Now()
	if err := m.UpdateWithVersion(session, data); err != nil {
		return errors.Wrapf(xerr.NewErrMsg("删除数据失败"), "ThirdPaymentModel delete err : %+v", err)
	}
	return nil
}

//暴露给logic开启事务
func (m *defaultThirdPaymentModel) Trans(fn func(session sqlx.Session) error) error {

	err := m.Transact(func(session sqlx.Session) error {
		return fn(session)
	})
	return err

}

//格式化缓存key
func (m *defaultThirdPaymentModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheLooklookPaymentThirdPaymentIdPrefix, primary)
}

//根据主键去db查询一条数据
func (m *defaultThirdPaymentModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? and del_state = ? limit 1", thirdPaymentRows, m.table)
	return conn.QueryRow(v, query, primary, globalkey.DelStateNo)
}

//!!!!! 其他自定义方法，从此处开始写,此处上方不要写自定义方法!!!!!
