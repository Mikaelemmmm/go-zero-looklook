package model

import (
	"context"
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
	homestayOrderFieldNames          = builder.RawFieldNames(&HomestayOrder{})
	homestayOrderRows                = strings.Join(homestayOrderFieldNames, ",")
	homestayOrderRowsExpectAutoSet   = strings.Join(stringx.Remove(homestayOrderFieldNames, "`id`", "`create_time`", "`update_time`"), ",")
	homestayOrderRowsWithPlaceHolder = strings.Join(stringx.Remove(homestayOrderFieldNames, "`id`", "`create_time`", "`update_time`"), "=?,") + "=?"

	cacheLooklookOrderHomestayOrderIdPrefix = "cache:looklookOrder:homestayOrder:id:"
	cacheLooklookOrderHomestayOrderSnPrefix = "cache:looklookOrder:homestayOrder:sn:"
)

type (
	HomestayOrderModel interface {
		//新增数据
		Insert(ctx context.Context, session sqlx.Session, data *HomestayOrder) (sql.Result, error)

		//根据主键查询一条数据，走缓存
		FindOne(ctx context.Context, id int64) (*HomestayOrder, error)

		//根据唯一索引查询一条数据，走缓存
		FindOneBySn(ctx context.Context, sn string) (*HomestayOrder, error)

		//删除数据
		Delete(ctx context.Context, session sqlx.Session, id int64) error

		//软删除数据
		DeleteSoft(ctx context.Context, session sqlx.Session, data *HomestayOrder) error

		//更新数据
		Update(ctx context.Context, session sqlx.Session, data *HomestayOrder) (sql.Result, error)

		//更新数据，使用乐观锁
		UpdateWithVersion(ctx context.Context, session sqlx.Session, data *HomestayOrder) error

		//根据条件查询一条数据，不走缓存
		FindOneByQuery(ctx context.Context, rowBuilder squirrel.SelectBuilder) (*HomestayOrder, error)

		//sum某个字段
		FindSum(ctx context.Context, sumBuilder squirrel.SelectBuilder) (float64, error)

		//根据条件统计条数
		FindCount(ctx context.Context, countBuilder squirrel.SelectBuilder) (int64, error)

		//查询所有数据不分页
		FindAll(ctx context.Context, rowBuilder squirrel.SelectBuilder, orderBy string) ([]*HomestayOrder, error)

		//根据页码分页查询分页数据
		FindPageListByPage(ctx context.Context, rowBuilder squirrel.SelectBuilder, page, pageSize int64, orderBy string) ([]*HomestayOrder, error)

		//根据id倒序分页查询分页数据
		FindPageListByIdDESC(ctx context.Context, rowBuilder squirrel.SelectBuilder, preMinId, pageSize int64) ([]*HomestayOrder, error)

		//根据id升序分页查询分页数据
		FindPageListByIdASC(ctx context.Context, rowBuilder squirrel.SelectBuilder, preMaxId, pageSize int64) ([]*HomestayOrder, error)

		//暴露给logic，开启事务
		Trans(ctx context.Context, fn func(context context.Context, session sqlx.Session) error) error

		//暴露给logic，查询数据的builder
		RowBuilder() squirrel.SelectBuilder

		//暴露给logic，查询count的builder
		CountBuilder(field string) squirrel.SelectBuilder

		//暴露给logic，查询sum的builder
		SumBuilder(field string) squirrel.SelectBuilder
	}

	defaultHomestayOrderModel struct {
		sqlc.CachedConn
		table string
	}

	HomestayOrder struct {
		Id                  int64     `db:"id"`
		CreateTime          time.Time `db:"create_time"`
		UpdateTime          time.Time `db:"update_time"`
		DeleteTime          time.Time `db:"delete_time"`
		DelState            int64     `db:"del_state"`
		Version             int64     `db:"version"`               // 版本号
		Sn                  string    `db:"sn"`                    // 订单号
		UserId              int64     `db:"user_id"`               // 下单用户id
		HomestayId          int64     `db:"homestay_id"`           // 民宿id
		Title               string    `db:"title"`                 // 标题
		SubTitle            string    `db:"sub_title"`             // 副标题
		Cover               string    `db:"cover"`                 // 封面
		Info                string    `db:"info"`                  // 介绍
		PeopleNum           int64     `db:"people_num"`            // 容纳人的数量
		RowType             int64     `db:"row_type"`              // 售卖类型0：按房间出售 1:按人次出售
		NeedFood            int64     `db:"need_food"`             // 0:不需要餐食 1:需要参数
		FoodInfo            string    `db:"food_info"`             // 餐食标准
		FoodPrice           int64     `db:"food_price"`            // 餐食价格(分)
		HomestayPrice       int64     `db:"homestay_price"`        // 民宿价格(分)
		MarketHomestayPrice int64     `db:"market_homestay_price"` // 民宿市场价格(分)
		HomestayBusinessId  int64     `db:"homestay_business_id"`  // 店铺id
		HomestayUserId      int64     `db:"homestay_user_id"`      // 店铺房东id
		LiveStartDate       time.Time `db:"live_start_date"`       // 开始入住日期
		LiveEndDate         time.Time `db:"live_end_date"`         // 结束入住日期
		LivePeopleNum       int64     `db:"live_people_num"`       // 实际入住人数
		TradeState          int64     `db:"trade_state"`           // -1: 已取消 0:待支付 1:未使用 2:已使用  3:已退款 4:已过期
		TradeCode           string    `db:"trade_code"`            // 确认码
		Remark              string    `db:"remark"`                // 用户下单备注
		OrderTotalPrice     int64     `db:"order_total_price"`     // 订单总价格（餐食总价格+民宿总价格）(分)
		FoodTotalPrice      int64     `db:"food_total_price"`      // 餐食总价格(分)
		HomestayTotalPrice  int64     `db:"homestay_total_price"`  // 民宿总价格(分)
	}
)

func NewHomestayOrderModel(conn sqlx.SqlConn, c cache.CacheConf) HomestayOrderModel {
	return &defaultHomestayOrderModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "`homestay_order`",
	}
}

func (m *defaultHomestayOrderModel) Insert(ctx context.Context, session sqlx.Session, data *HomestayOrder) (sql.Result, error) {
	data.DeleteTime = time.Unix(0, 0)
	looklookOrderHomestayOrderIdKey := fmt.Sprintf("%s%v", cacheLooklookOrderHomestayOrderIdPrefix, data.Id)
	looklookOrderHomestayOrderSnKey := fmt.Sprintf("%s%v", cacheLooklookOrderHomestayOrderSnPrefix, data.Sn)
	return m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, homestayOrderRowsExpectAutoSet)
		if session != nil {
			return session.ExecCtx(ctx, query, data.DeleteTime, data.DelState, data.Version, data.Sn, data.UserId, data.HomestayId, data.Title, data.SubTitle, data.Cover, data.Info, data.PeopleNum, data.RowType, data.NeedFood, data.FoodInfo, data.FoodPrice, data.HomestayPrice, data.MarketHomestayPrice, data.HomestayBusinessId, data.HomestayUserId, data.LiveStartDate, data.LiveEndDate, data.LivePeopleNum, data.TradeState, data.TradeCode, data.Remark, data.OrderTotalPrice, data.FoodTotalPrice, data.HomestayTotalPrice)
		}
		return conn.ExecCtx(ctx, query, data.DeleteTime, data.DelState, data.Version, data.Sn, data.UserId, data.HomestayId, data.Title, data.SubTitle, data.Cover, data.Info, data.PeopleNum, data.RowType, data.NeedFood, data.FoodInfo, data.FoodPrice, data.HomestayPrice, data.MarketHomestayPrice, data.HomestayBusinessId, data.HomestayUserId, data.LiveStartDate, data.LiveEndDate, data.LivePeopleNum, data.TradeState, data.TradeCode, data.Remark, data.OrderTotalPrice, data.FoodTotalPrice, data.HomestayTotalPrice)
	}, looklookOrderHomestayOrderIdKey, looklookOrderHomestayOrderSnKey)
}

//根据主键查询一条数据，走缓存
func (m *defaultHomestayOrderModel) FindOne(ctx context.Context, id int64) (*HomestayOrder, error) {
	looklookOrderHomestayOrderIdKey := fmt.Sprintf("%s%v", cacheLooklookOrderHomestayOrderIdPrefix, id)
	var resp HomestayOrder
	err := m.QueryRowCtx(ctx, &resp, looklookOrderHomestayOrderIdKey, func(ctx context.Context, conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? and del_state = ? limit 1", homestayOrderRows, m.table)
		return conn.QueryRowCtx(ctx, v, query, id, globalkey.DelStateNo)
	})
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

//根据唯一索引查询一条数据，走缓存
func (m *defaultHomestayOrderModel) FindOneBySn(ctx context.Context, sn string) (*HomestayOrder, error) {
	looklookOrderHomestayOrderSnKey := fmt.Sprintf("%s%v", cacheLooklookOrderHomestayOrderSnPrefix, sn)
	var resp HomestayOrder
	err := m.QueryRowIndexCtx(ctx, &resp, looklookOrderHomestayOrderSnKey, m.formatPrimary, func(ctx context.Context, conn sqlx.SqlConn, v interface{}) (i interface{}, e error) {
		query := fmt.Sprintf("select %s from %s where `sn` = ? and del_state = ? limit 1", homestayOrderRows, m.table)
		if err := conn.QueryRowCtx(ctx, &resp, query, sn, globalkey.DelStateNo); err != nil {
			return nil, err
		}
		return resp.Id, nil
	}, m.queryPrimary)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultHomestayOrderModel) Delete(ctx context.Context, session sqlx.Session, id int64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}

	looklookOrderHomestayOrderIdKey := fmt.Sprintf("%s%v", cacheLooklookOrderHomestayOrderIdPrefix, id)
	looklookOrderHomestayOrderSnKey := fmt.Sprintf("%s%v", cacheLooklookOrderHomestayOrderSnPrefix, data.Sn)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
		if session != nil {
			return session.ExecCtx(ctx, query, id)
		}
		return conn.ExecCtx(ctx, query, id)
	}, looklookOrderHomestayOrderIdKey, looklookOrderHomestayOrderSnKey)
	return err
}

//软删除数据
func (m *defaultHomestayOrderModel) DeleteSoft(ctx context.Context, session sqlx.Session, data *HomestayOrder) error {
	data.DelState = globalkey.DelStateYes
	data.DeleteTime = time.Now()
	if err := m.UpdateWithVersion(ctx, session, data); err != nil {
		return errors.Wrapf(xerr.NewErrMsg("删除数据失败"), "HomestayOrderModel delete err : %+v", err)
	}
	return nil
}

//暴露给logic开启事务
func (m *defaultHomestayOrderModel) Trans(ctx context.Context, fn func(ctx context.Context, session sqlx.Session) error) error {

	return m.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		return fn(ctx, session)
	})

}

func (m *defaultHomestayOrderModel) Update(ctx context.Context, session sqlx.Session, data *HomestayOrder) (sql.Result, error) {
	looklookOrderHomestayOrderIdKey := fmt.Sprintf("%s%v", cacheLooklookOrderHomestayOrderIdPrefix, data.Id)
	looklookOrderHomestayOrderSnKey := fmt.Sprintf("%s%v", cacheLooklookOrderHomestayOrderSnPrefix, data.Sn)
	return m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, homestayOrderRowsWithPlaceHolder)
		if session != nil {
			return session.ExecCtx(ctx, query, data.DeleteTime, data.DelState, data.Version, data.Sn, data.UserId, data.HomestayId, data.Title, data.SubTitle, data.Cover, data.Info, data.PeopleNum, data.RowType, data.NeedFood, data.FoodInfo, data.FoodPrice, data.HomestayPrice, data.MarketHomestayPrice, data.HomestayBusinessId, data.HomestayUserId, data.LiveStartDate, data.LiveEndDate, data.LivePeopleNum, data.TradeState, data.TradeCode, data.Remark, data.OrderTotalPrice, data.FoodTotalPrice, data.HomestayTotalPrice, data.Id)
		}
		return conn.ExecCtx(ctx, query, data.DeleteTime, data.DelState, data.Version, data.Sn, data.UserId, data.HomestayId, data.Title, data.SubTitle, data.Cover, data.Info, data.PeopleNum, data.RowType, data.NeedFood, data.FoodInfo, data.FoodPrice, data.HomestayPrice, data.MarketHomestayPrice, data.HomestayBusinessId, data.HomestayUserId, data.LiveStartDate, data.LiveEndDate, data.LivePeopleNum, data.TradeState, data.TradeCode, data.Remark, data.OrderTotalPrice, data.FoodTotalPrice, data.HomestayTotalPrice, data.Id)
	}, looklookOrderHomestayOrderIdKey, looklookOrderHomestayOrderSnKey)
}

//乐观锁修改数据 ,推荐使用
func (m *defaultHomestayOrderModel) UpdateWithVersion(ctx context.Context, session sqlx.Session, data *HomestayOrder) error {

	oldVersion := data.Version
	data.Version += 1

	var sqlResult sql.Result
	var err error

	looklookOrderHomestayOrderIdKey := fmt.Sprintf("%s%v", cacheLooklookOrderHomestayOrderIdPrefix, data.Id)
	looklookOrderHomestayOrderSnKey := fmt.Sprintf("%s%v", cacheLooklookOrderHomestayOrderSnPrefix, data.Sn)
	sqlResult, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ? and version = ? ", m.table, homestayOrderRowsWithPlaceHolder)
		if session != nil {
			return session.ExecCtx(ctx, query, data.DeleteTime, data.DelState, data.Version, data.Sn, data.UserId, data.HomestayId, data.Title, data.SubTitle, data.Cover, data.Info, data.PeopleNum, data.RowType, data.NeedFood, data.FoodInfo, data.FoodPrice, data.HomestayPrice, data.MarketHomestayPrice, data.HomestayBusinessId, data.HomestayUserId, data.LiveStartDate, data.LiveEndDate, data.LivePeopleNum, data.TradeState, data.TradeCode, data.Remark, data.OrderTotalPrice, data.FoodTotalPrice, data.HomestayTotalPrice, data.Id, oldVersion)
		}
		return conn.ExecCtx(ctx, query, data.DeleteTime, data.DelState, data.Version, data.Sn, data.UserId, data.HomestayId, data.Title, data.SubTitle, data.Cover, data.Info, data.PeopleNum, data.RowType, data.NeedFood, data.FoodInfo, data.FoodPrice, data.HomestayPrice, data.MarketHomestayPrice, data.HomestayBusinessId, data.HomestayUserId, data.LiveStartDate, data.LiveEndDate, data.LivePeopleNum, data.TradeState, data.TradeCode, data.Remark, data.OrderTotalPrice, data.FoodTotalPrice, data.HomestayTotalPrice, data.Id, oldVersion)
	}, looklookOrderHomestayOrderIdKey, looklookOrderHomestayOrderSnKey)
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
func (m *defaultHomestayOrderModel) FindOneByQuery(ctx context.Context, rowBuilder squirrel.SelectBuilder) (*HomestayOrder, error) {

	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).ToSql()
	if err != nil {
		return nil, err
	}

	var resp HomestayOrder
	err = m.QueryRowNoCacheCtx(ctx, &resp, query, values...)
	switch err {
	case nil:
		return &resp, nil
	default:
		return nil, err
	}
}

//统计某个字段总和
func (m *defaultHomestayOrderModel) FindSum(ctx context.Context, sumBuilder squirrel.SelectBuilder) (float64, error) {

	query, values, err := sumBuilder.Where("del_state = ?", globalkey.DelStateNo).ToSql()
	if err != nil {
		return 0, err
	}

	var resp float64
	err = m.QueryRowNoCacheCtx(ctx, &resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return 0, err
	}
}

//根据某个字段查询数据数量
func (m *defaultHomestayOrderModel) FindCount(ctx context.Context, countBuilder squirrel.SelectBuilder) (int64, error) {

	query, values, err := countBuilder.Where("del_state = ?", globalkey.DelStateNo).ToSql()
	if err != nil {
		return 0, err
	}

	var resp int64
	err = m.QueryRowNoCacheCtx(ctx, &resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return 0, err
	}
}

//查询所有数据
func (m *defaultHomestayOrderModel) FindAll(ctx context.Context, rowBuilder squirrel.SelectBuilder, orderBy string) ([]*HomestayOrder, error) {

	if orderBy == "" {
		rowBuilder = rowBuilder.OrderBy("id DESC")
	} else {
		rowBuilder = rowBuilder.OrderBy(orderBy)
	}

	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).ToSql()
	if err != nil {
		return nil, err
	}

	var resp []*HomestayOrder
	err = m.QueryRowsNoCacheCtx(ctx, &resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

//按照页码分页查询数据
func (m *defaultHomestayOrderModel) FindPageListByPage(ctx context.Context, rowBuilder squirrel.SelectBuilder, page, pageSize int64, orderBy string) ([]*HomestayOrder, error) {

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

	var resp []*HomestayOrder
	err = m.QueryRowsNoCacheCtx(ctx, &resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

//按照id倒序分页查询数据，不支持排序
func (m *defaultHomestayOrderModel) FindPageListByIdDESC(ctx context.Context, rowBuilder squirrel.SelectBuilder, preMinId, pageSize int64) ([]*HomestayOrder, error) {

	if preMinId > 0 {
		rowBuilder = rowBuilder.Where(" id < ? ", preMinId)
	}

	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).OrderBy("id DESC").Limit(uint64(pageSize)).ToSql()
	if err != nil {
		return nil, err
	}

	var resp []*HomestayOrder
	err = m.QueryRowsNoCacheCtx(ctx, &resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

//按照id升序分页查询数据，不支持排序
func (m *defaultHomestayOrderModel) FindPageListByIdASC(ctx context.Context, rowBuilder squirrel.SelectBuilder, preMaxId, pageSize int64) ([]*HomestayOrder, error) {

	if preMaxId > 0 {
		rowBuilder = rowBuilder.Where(" id > ? ", preMaxId)
	}

	query, values, err := rowBuilder.Where("del_state = ?", globalkey.DelStateNo).OrderBy("id ASC").Limit(uint64(pageSize)).ToSql()
	if err != nil {
		return nil, err
	}

	var resp []*HomestayOrder
	err = m.QueryRowsNoCacheCtx(ctx, &resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

//暴露给logic查询数据构建条件使用的builder
func (m *defaultHomestayOrderModel) RowBuilder() squirrel.SelectBuilder {
	return squirrel.Select(homestayOrderRows).From(m.table)
}

//暴露给logic查询count构建条件使用的builder
func (m *defaultHomestayOrderModel) CountBuilder(field string) squirrel.SelectBuilder {
	return squirrel.Select("COUNT(" + field + ")").From(m.table)
}

//暴露给logic查询构建条件使用的builder
func (m *defaultHomestayOrderModel) SumBuilder(field string) squirrel.SelectBuilder {
	return squirrel.Select("IFNULL(SUM(" + field + "),0)").From(m.table)
}

//格式化缓存key
func (m *defaultHomestayOrderModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheLooklookOrderHomestayOrderIdPrefix, primary)
}

//根据主键去db查询一条数据
func (m *defaultHomestayOrderModel) queryPrimary(ctx context.Context, conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? and del_state = ? limit 1", homestayOrderRows, m.table)
	return conn.QueryRowCtx(ctx, v, query, primary, globalkey.DelStateNo)
}

//----------------------------------------其他自定义方法，从此处开始写,此处上方不要写自定义方法----------------------------------------
