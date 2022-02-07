package model

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"looklook/common/globalkey"

	sqlBuilder "github.com/didi/gendry/builder"
	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	homestayOrderFieldNames          = builder.RawFieldNames(&HomestayOrder{})
	homestayOrderRows                = strings.Join(homestayOrderFieldNames, ",")
	homestayOrderRowsWithPlaceHolder = strings.Join(stringx.Remove(homestayOrderFieldNames, "`id`", "`create_time`", "`update_time`", "`version`"), "=?,") + "=?"

	cacheLooklookOrderHomestayOrderIdPrefix = "cache:looklookOrder:homestayOrder:id:"
	cacheLooklookOrderHomestayOrderSnPrefix = "cache:looklookOrder:homestayOrder:sn:"
)

type (
	HomestayOrderModel interface {
		ListByUserIdTradeState(lastId, pageSize, userId int64, tradeState int64) ([]*HomestayOrder, error)
		FindOne(id int64) (*HomestayOrder, error)
		FindOneBySn(sn string) (*HomestayOrder, error)
		Insert(session sqlx.Session, data *HomestayOrder) (sql.Result, error)
		Update(session sqlx.Session, data *HomestayOrder) error
		Delete(session sqlx.Session, data *HomestayOrder) error
		Trans(fn func(session sqlx.Session) error) error
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
		FoodInfo            string    `db:"food_info"`             // 餐食标准(分)
		FoodPrice           int64     `db:"food_price"`            // 餐食价格(分)
		HomestayPrice       int64     `db:"homestay_price"`        // 民宿价格(分)
		MarketHomestayPrice int64     `db:"market_homestay_price"` // 民宿市场价格
		HomestayBusinessId  int64     `db:"homestay_business_id"`  // 店铺id
		HomestayUserId      int64     `db:"homestay_user_id"`      // 店铺房东id
		LiveStartDate       time.Time `db:"live_start_date"`       // 开始入住日期
		LiveEndDate         time.Time `db:"live_end_date"`         // 结束入住日期
		LivePeopleNum       int64     `db:"live_people_num"`       // 实际入住人数
		TradeState          int64     `db:"trade_state"`           // -1: 已取消 0:待支付 1:未使用 2:已使用  3:已退款 4:已过期
		TradeCode           string    `db:"trade_code"`            // 确认码
		Remark              string    `db:"remark"`                // 用户下单备注
		OrderTotalPrice     int64     `db:"order_total_price"`     // 订单总价格（餐食总价格+民宿总价格）（分）
		FoodTotalPrice      int64     `db:"food_total_price"`      // 餐食总价格（分）
		HomestayTotalPrice  int64     `db:"homestay_total_price"`  // 民宿总价格（分）
	}
)

func NewHomestayOrderModel(conn sqlx.SqlConn, c cache.CacheConf) HomestayOrderModel {
	return &defaultHomestayOrderModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "`homestay_order`",
	}
}

func (m *defaultHomestayOrderModel) ListByUserIdTradeState(lastId, pageSize, userId int64, tradeState int64) ([]*HomestayOrder, error) {

	if lastId == 0 {
		lastId = math.MaxInt64
	}

	where := map[string]interface{}{
		"`user_id`":   userId,
		"`del_state`": globalkey.DelStateNo,
		"`id` <":      lastId,
		"_orderby":    "id DESC",
		"_limit":      []uint{0, uint(pageSize)},
	}

	//有支持的状态在筛选，否则返回所有
	if tradeState >= HomestayOrderTradeStateCancel && tradeState <= HomestayOrderTradeStateExpire {
		where["`trade_state`"] = tradeState
	}

	query, values, err := sqlBuilder.BuildSelect(m.table, where, homestayOrderFieldNames)
	if err != nil {
		return nil, err
	}

	var resp []*HomestayOrder
	err = m.QueryRowsNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

func (m *defaultHomestayOrderModel) Insert(session sqlx.Session, data *HomestayOrder) (sql.Result, error) {

	query := fmt.Sprintf("insert into  %s (sn,user_id,homestay_id,title,sub_title,cover,info,people_num,row_type,need_food,food_info,food_price,homestay_price,market_homestay_price,homestay_business_id,homestay_user_id,live_start_date,live_end_date,live_people_num,trade_state,trade_code,remark,order_total_price,food_total_price,homestay_total_price) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", m.table)
	if session != nil {
		return session.Exec(query, data.Sn, data.UserId, data.HomestayId, data.Title, data.SubTitle, data.Cover, data.Info,
			data.PeopleNum, data.RowType, data.NeedFood, data.FoodInfo, data.FoodPrice, data.HomestayPrice, data.MarketHomestayPrice,
			data.HomestayBusinessId, data.HomestayUserId,
			data.LiveStartDate, data.LiveEndDate, data.LivePeopleNum, data.TradeState, data.TradeCode, data.Remark,
			data.OrderTotalPrice, data.FoodTotalPrice, data.HomestayTotalPrice)
	}
	return m.ExecNoCache(query, data.Sn, data.UserId, data.HomestayId, data.Title, data.SubTitle, data.Cover, data.Info,
		data.PeopleNum, data.RowType, data.NeedFood, data.FoodInfo, data.FoodPrice, data.HomestayPrice, data.MarketHomestayPrice,
		data.HomestayBusinessId, data.HomestayUserId,
		data.LiveStartDate, data.LiveEndDate, data.LivePeopleNum, data.TradeState, data.TradeCode, data.Remark,
		data.OrderTotalPrice, data.FoodTotalPrice, data.HomestayTotalPrice)

}

func (m *defaultHomestayOrderModel) FindOne(id int64) (*HomestayOrder, error) {
	looklookOrderHomestayOrderIdKey := fmt.Sprintf("%s%v", cacheLooklookOrderHomestayOrderIdPrefix, id)
	var resp HomestayOrder
	err := m.QueryRow(&resp, looklookOrderHomestayOrderIdKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", homestayOrderRows, m.table)
		return conn.QueryRow(v, query, id)
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

func (m *defaultHomestayOrderModel) FindOneBySn(sn string) (*HomestayOrder, error) {
	looklookOrderHomestayOrderSnKey := fmt.Sprintf("%s%v", cacheLooklookOrderHomestayOrderSnPrefix, sn)
	var resp HomestayOrder
	err := m.QueryRowIndex(&resp, looklookOrderHomestayOrderSnKey, m.formatPrimary, func(conn sqlx.SqlConn, v interface{}) (i interface{}, e error) {
		query := fmt.Sprintf("select %s from %s where `sn` = ? limit 1", homestayOrderRows, m.table)
		if err := conn.QueryRow(&resp, query, sn); err != nil {
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

func (m *defaultHomestayOrderModel) Update(session sqlx.Session, data *HomestayOrder) error {
	looklookOrderHomestayOrderIdKey := fmt.Sprintf("%s%v", cacheLooklookOrderHomestayOrderIdPrefix, data.Id)
	looklookOrderHomestayOrderSnKey := fmt.Sprintf("%s%v", cacheLooklookOrderHomestayOrderSnPrefix, data.Sn)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, homestayOrderRowsWithPlaceHolder)
		if session != nil {
			return session.Exec(query, data.DeleteTime, data.DelState, data.Sn, data.UserId, data.HomestayId, data.Title,
				data.SubTitle, data.Cover, data.Info, data.PeopleNum, data.RowType, data.NeedFood, data.FoodInfo, data.FoodPrice, data.HomestayPrice,
				data.MarketHomestayPrice, data.HomestayBusinessId, data.HomestayUserId, data.LiveStartDate, data.LiveEndDate,
				data.LivePeopleNum, data.TradeState, data.TradeCode, data.Remark, data.OrderTotalPrice, data.FoodTotalPrice, data.HomestayTotalPrice, data.Id)
		}
		return conn.Exec(query, data.DeleteTime, data.DelState, data.Sn, data.UserId, data.HomestayId, data.Title, data.SubTitle,
			data.Cover, data.Info, data.PeopleNum, data.RowType, data.NeedFood, data.FoodInfo, data.FoodPrice, data.HomestayPrice, data.MarketHomestayPrice,
			data.HomestayBusinessId, data.HomestayUserId, data.LiveStartDate, data.LiveEndDate, data.LivePeopleNum, data.TradeState,
			data.TradeCode, data.Remark, data.OrderTotalPrice, data.FoodTotalPrice, data.HomestayTotalPrice, data.Id)
	}, looklookOrderHomestayOrderIdKey, looklookOrderHomestayOrderSnKey)
	return err
}

func (m *defaultHomestayOrderModel) Delete(session sqlx.Session, data *HomestayOrder) error {
	data.DelState = globalkey.DelStateYes
	return m.Update(session, data)
}

func (m *defaultHomestayOrderModel) Trans(fn func(session sqlx.Session) error) error {

	err := m.Transact(func(session sqlx.Session) error {
		return fn(session)
	})
	return err

}

func (m *defaultHomestayOrderModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheLooklookOrderHomestayOrderIdPrefix, primary)
}

func (m *defaultHomestayOrderModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", homestayOrderRows, m.table)
	return conn.QueryRow(v, query, primary)
}
