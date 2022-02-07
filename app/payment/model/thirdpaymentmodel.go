package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"looklook/common/globalkey"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	thirdPaymentFieldNames          = builder.RawFieldNames(&ThirdPayment{})
	thirdPaymentRows                = strings.Join(thirdPaymentFieldNames, ",")
	thirdPaymentRowsWithPlaceHolder = strings.Join(stringx.Remove(thirdPaymentFieldNames, "`id`", "`create_time`", "`update_time`", "`version`"), "=?,") + "=?"

	cacheLooklookPaymentThirdPaymentIdPrefix = "cache:looklookPayment:thirdPayment:id:"
	cacheLooklookPaymentThirdPaymentSnPrefix = "cache:looklookPayment:thirdPayment:sn:"
)

type (
	ThirdPaymentModel interface {
		FindOnePaySucessOrRefundByOrderSn(orderSn string) (*ThirdPayment, error)
		FindOne(id int64) (*ThirdPayment, error)
		FindOneBySn(sn string) (*ThirdPayment, error)
		Insert(session sqlx.Session, data *ThirdPayment) (sql.Result, error)
		Update(session sqlx.Session, data *ThirdPayment) error
		Delete(session sqlx.Session, data *ThirdPayment) error
		Trans(fn func(session sqlx.Session) error) error
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
		PayTotal       int64     `db:"pay_total"`        // 支付总金额
		TransactionId  string    `db:"transaction_id"`   // 第三方支付单号
		TradeStateDesc string    `db:"trade_state_desc"` // 支付状态描述
		OrderSn        string    `db:"order_sn"`         // 业务单号
		ServiceType    string    `db:"service_type"`     // 业务类型
		PayStatus      int64     `db:"pay_status"`       // 平台内交易状态  0:未支付 1:支付成功 2:已退款 -1:支付失败
		PayTime        time.Time `db:"pay_time"`         // 支付成功时间
	}
)

func NewThirdPaymentModel(conn sqlx.SqlConn, c cache.CacheConf) ThirdPaymentModel {
	return &defaultThirdPaymentModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "`third_payment`",
	}
}

//查询成功的订单
func (m *defaultThirdPaymentModel) FindOnePaySucessOrRefundByOrderSn(orderSn string) (*ThirdPayment, error) {

	var resp ThirdPayment
	query := fmt.Sprintf("select %s from %s where `order_sn` = ? and (trade_state = ? or trade_state = ? ) limit 1", thirdPaymentRows, m.table)
	err := m.CachedConn.QueryRowNoCache(&resp, query, orderSn, ThirdPaymentPayTradeStateSuccess, ThirdPaymentPayTradeStateRefund)

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

func (m *defaultThirdPaymentModel) Insert(session sqlx.Session, data *ThirdPayment) (sql.Result, error) {
	looklookPaymentThirdPaymentIdKey := fmt.Sprintf("%s%v", cacheLooklookPaymentThirdPaymentIdPrefix, data.Id)
	looklookPaymentThirdPaymentSnKey := fmt.Sprintf("%s%v", cacheLooklookPaymentThirdPaymentSnPrefix, data.Sn)
	return m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {

		query := fmt.Sprintf("insert into  %s(sn,user_id,pay_mode,pay_total,transaction_id,order_sn,service_type) values(?,?,?,?,?,?,?)", m.table)
		if session != nil {
			return session.Exec(query, data.Sn, data.UserId, data.PayMode, data.PayTotal, data.TransactionId, data.OrderSn, data.ServiceType)
		}
		return m.ExecNoCache(query, data.Sn, data.UserId, data.PayMode, data.PayTotal, data.TransactionId, data.OrderSn, data.ServiceType)
	}, looklookPaymentThirdPaymentIdKey, looklookPaymentThirdPaymentSnKey)

}

func (m *defaultThirdPaymentModel) FindOne(id int64) (*ThirdPayment, error) {
	looklookPaymentThirdPaymentIdKey := fmt.Sprintf("%s%v", cacheLooklookPaymentThirdPaymentIdPrefix, id)
	var resp ThirdPayment
	err := m.QueryRow(&resp, looklookPaymentThirdPaymentIdKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", thirdPaymentRows, m.table)
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

func (m *defaultThirdPaymentModel) FindOneBySn(sn string) (*ThirdPayment, error) {
	looklookPaymentThirdPaymentSnKey := fmt.Sprintf("%s%v", cacheLooklookPaymentThirdPaymentSnPrefix, sn)
	var resp ThirdPayment
	err := m.QueryRowIndex(&resp, looklookPaymentThirdPaymentSnKey, m.formatPrimary, func(conn sqlx.SqlConn, v interface{}) (i interface{}, e error) {
		query := fmt.Sprintf("select %s from %s where `sn` = ? limit 1", thirdPaymentRows, m.table)
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

func (m *defaultThirdPaymentModel) Update(session sqlx.Session, data *ThirdPayment) error {
	looklookPaymentThirdPaymentIdKey := fmt.Sprintf("%s%v", cacheLooklookPaymentThirdPaymentIdPrefix, data.Id)
	looklookPaymentThirdPaymentSnKey := fmt.Sprintf("%s%v", cacheLooklookPaymentThirdPaymentSnPrefix, data.Sn)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s , `version` = `version` + 1 where `id` = ?  and version = ?", m.table, thirdPaymentRowsWithPlaceHolder)
		if session != nil {
			return session.Exec(query, data.Sn, data.DeleteTime, data.DelState, data.UserId, data.PayMode, data.TradeType, data.TradeState, data.PayTotal, data.TransactionId, data.TradeStateDesc, data.OrderSn, data.ServiceType, data.PayStatus, data.PayTime, data.Id, data.Version)
		}
		return conn.Exec(query, data.Sn, data.DeleteTime, data.DelState, data.UserId, data.PayMode, data.TradeType, data.TradeState, data.PayTotal, data.TransactionId, data.TradeStateDesc, data.OrderSn, data.ServiceType, data.PayStatus, data.PayTime, data.Id, data.Version)
	}, looklookPaymentThirdPaymentIdKey, looklookPaymentThirdPaymentSnKey)
	return err
}

func (m *defaultThirdPaymentModel) Delete(session sqlx.Session, data *ThirdPayment) error {
	data.DelState = globalkey.DelStateYes
	return m.Update(session, data)
}

func (m *defaultThirdPaymentModel) Trans(fn func(session sqlx.Session) error) error {

	err := m.Transact(func(session sqlx.Session) error {
		return fn(session)
	})
	return err

}

func (m *defaultThirdPaymentModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheLooklookPaymentThirdPaymentIdPrefix, primary)
}

func (m *defaultThirdPaymentModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", thirdPaymentRows, m.table)
	return conn.QueryRow(v, query, primary)
}
