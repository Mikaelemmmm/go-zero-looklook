package model

import (
	"database/sql"
	"fmt"
	"math"
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
	homestayActivityFieldNames          = builder.RawFieldNames(&HomestayActivity{})
	homestayActivityRows                = strings.Join(homestayActivityFieldNames, ",")
	homestayActivityRowsWithPlaceHolder = strings.Join(stringx.Remove(homestayActivityFieldNames, "`id`", "`create_time`", "`update_time`", "`version`"), "=?,") + "=?"

	cacheLooklookTravelHomestayActivityIdPrefix = "cache:looklookTravel:homestayActivity:id:"
)

type (
	HomestayActivityModel interface {
		FindPageByRowTypeStatus(lastId, pageSize int64, rowType string, rowStatus int64) ([]int64, error)
		FindOne(id int64) (*HomestayActivity, error)
		Insert(session sqlx.Session, data *HomestayActivity) (sql.Result, error)
		Update(session sqlx.Session, data *HomestayActivity) error
		Delete(session sqlx.Session, data *HomestayActivity) error
		Trans(fn func(session sqlx.Session) error) error
	}

	defaultHomestayActivityModel struct {
		sqlc.CachedConn
		table string
	}

	HomestayActivity struct {
		Id         int64     `db:"id"`
		CreateTime time.Time `db:"create_time"`
		UpdateTime time.Time `db:"update_time"`
		DeleteTime time.Time `db:"delete_time"`
		DelState   int64     `db:"del_state"`
		RowType    string    `db:"row_type"`   // 活动类型
		DataId     int64     `db:"data_id"`    // 业务表id（id跟随活动类型走）
		RowStatus  int64     `db:"row_status"` // 0:下架 1:上架
	}
)

func NewHomestayActivityModel(conn sqlx.SqlConn, c cache.CacheConf) HomestayActivityModel {
	return &defaultHomestayActivityModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "`homestay_activity`",
	}
}

func (m *defaultHomestayActivityModel) FindPageByRowTypeStatus(lastId, pageSize int64, rowType string, rowStatus int64) ([]int64, error) {

	if lastId == 0 {
		lastId = math.MaxInt64
	}

	var resp []int64
	query := fmt.Sprintf("select data_id from %s where row_type = ? and row_status = ? and del_state = ? and data_id < ? order by data_id desc limit ?", m.table)
	err := m.QueryRowsNoCache(&resp, query, rowType, rowStatus, globalkey.DelStateNo, lastId, pageSize)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

func (m *defaultHomestayActivityModel) Insert(session sqlx.Session, data *HomestayActivity) (sql.Result, error) {

	//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
	query := fmt.Sprintf("insert into .... (%s) values ...", m.table)
	if session != nil {
		//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
		return session.Exec(query, data.DeleteTime, data.DelState, data.RowType, data.DataId, data.RowStatus)
	}
	//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
	return m.ExecNoCache(query, data.DeleteTime, data.DelState, data.RowType, data.DataId, data.RowStatus)

}

func (m *defaultHomestayActivityModel) FindOne(id int64) (*HomestayActivity, error) {
	looklookTravelHomestayActivityIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayActivityIdPrefix, id)
	var resp HomestayActivity
	err := m.QueryRow(&resp, looklookTravelHomestayActivityIdKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", homestayActivityRows, m.table)
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

func (m *defaultHomestayActivityModel) Update(session sqlx.Session, data *HomestayActivity) error {
	looklookTravelHomestayActivityIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayActivityIdPrefix, data.Id)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, homestayActivityRowsWithPlaceHolder)
		if session != nil {
			return session.Exec(query, data.DeleteTime, data.DelState, data.RowType, data.DataId, data.RowStatus, data.Id)
		}
		return conn.Exec(query, data.DeleteTime, data.DelState, data.RowType, data.DataId, data.RowStatus, data.Id)
	}, looklookTravelHomestayActivityIdKey)
	return err
}

func (m *defaultHomestayActivityModel) Delete(session sqlx.Session, data *HomestayActivity) error {
	data.DelState = globalkey.DelStateYes
	return m.Update(session, data)
}

func (m *defaultHomestayActivityModel) Trans(fn func(session sqlx.Session) error) error {

	err := m.Transact(func(session sqlx.Session) error {
		return fn(session)
	})
	return err

}

func (m *defaultHomestayActivityModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheLooklookTravelHomestayActivityIdPrefix, primary)
}

func (m *defaultHomestayActivityModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", homestayActivityRows, m.table)
	return conn.QueryRow(v, query, primary)
}
