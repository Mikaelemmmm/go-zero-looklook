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
	homestayBusinessFieldNames          = builder.RawFieldNames(&HomestayBusiness{})
	homestayBusinessRows                = strings.Join(homestayBusinessFieldNames, ",")
	homestayBusinessRowsWithPlaceHolder = strings.Join(stringx.Remove(homestayBusinessFieldNames, "`id`", "`create_time`", "`update_time`", "`version`"), "=?,") + "=?"

	cacheLooklookTravelHomestayBusinessIdPrefix     = "cache:looklookTravel:homestayBusiness:id:"
	cacheLooklookTravelHomestayBusinessUserIdPrefix = "cache:looklookTravel:homestayBusiness:userId:"
)

type (
	HomestayBusinessModel interface {
		FindPageList(lastId, pageSize int64) ([]*HomestayBusiness, error)
		FindOne(id int64) (*HomestayBusiness, error)
		FindOneByUserId(userId int64) (*HomestayBusiness, error)
		Insert(session sqlx.Session, data *HomestayBusiness) (sql.Result, error)
		Update(session sqlx.Session, data *HomestayBusiness) error
		Delete(session sqlx.Session, data *HomestayBusiness) error
		Trans(fn func(session sqlx.Session) error) error
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
	}
)

func NewHomestayBusinessModel(conn sqlx.SqlConn, c cache.CacheConf) HomestayBusinessModel {
	return &defaultHomestayBusinessModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "`homestay_business`",
	}
}

func (m *defaultHomestayBusinessModel) FindPageList(lastId, pageSize int64) ([]*HomestayBusiness, error) {

	if lastId == 0 {
		lastId = math.MaxInt64
	}

	where := map[string]interface{}{
		"`del_state`": globalkey.DelStateNo,
		"`id` <":      lastId,
		"_orderby":    "id DESC",
		"_limit":      []uint{0, uint(pageSize)},
	}
	query, values, err := sqlBuilder.BuildSelect(m.table, where, homestayBusinessFieldNames)
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

func (m *defaultHomestayBusinessModel) Insert(session sqlx.Session, data *HomestayBusiness) (sql.Result, error) {
	looklookTravelHomestayBusinessIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayBusinessIdPrefix, data.Id)
	looklookTravelHomestayBusinessUserIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayBusinessUserIdPrefix, data.UserId)
	return m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {

		//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
		query := fmt.Sprintf("insert into .... (%s) values ...", m.table)
		if session != nil {
			//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
			return session.Exec(query, data.DeleteTime, data.DelState, data.Title, data.UserId, data.Info, data.BossInfo, data.LicenseFron, data.LicenseBack, data.RowState, data.Star, data.Tags, data.Cover, data.HeaderImg)
		}
		//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
		return conn.Exec(query, data.DeleteTime, data.DelState, data.Title, data.UserId, data.Info, data.BossInfo, data.LicenseFron, data.LicenseBack, data.RowState, data.Star, data.Tags, data.Cover, data.HeaderImg)
	}, looklookTravelHomestayBusinessUserIdKey, looklookTravelHomestayBusinessIdKey)

}

func (m *defaultHomestayBusinessModel) FindOne(id int64) (*HomestayBusiness, error) {
	looklookTravelHomestayBusinessIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayBusinessIdPrefix, id)
	var resp HomestayBusiness
	err := m.QueryRow(&resp, looklookTravelHomestayBusinessIdKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", homestayBusinessRows, m.table)
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

func (m *defaultHomestayBusinessModel) FindOneByUserId(userId int64) (*HomestayBusiness, error) {
	looklookTravelHomestayBusinessUserIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayBusinessUserIdPrefix, userId)
	var resp HomestayBusiness
	err := m.QueryRowIndex(&resp, looklookTravelHomestayBusinessUserIdKey, m.formatPrimary, func(conn sqlx.SqlConn, v interface{}) (i interface{}, e error) {
		query := fmt.Sprintf("select %s from %s where `user_id` = ? limit 1", homestayBusinessRows, m.table)
		if err := conn.QueryRow(&resp, query, userId); err != nil {
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

func (m *defaultHomestayBusinessModel) Update(session sqlx.Session, data *HomestayBusiness) error {
	looklookTravelHomestayBusinessIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayBusinessIdPrefix, data.Id)
	looklookTravelHomestayBusinessUserIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayBusinessUserIdPrefix, data.UserId)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, homestayBusinessRowsWithPlaceHolder)
		if session != nil {
			return session.Exec(query, data.DeleteTime, data.DelState, data.Title, data.UserId, data.Info, data.BossInfo, data.LicenseFron, data.LicenseBack, data.RowState, data.Star, data.Tags, data.Cover, data.HeaderImg, data.Id)
		}
		return conn.Exec(query, data.DeleteTime, data.DelState, data.Title, data.UserId, data.Info, data.BossInfo, data.LicenseFron, data.LicenseBack, data.RowState, data.Star, data.Tags, data.Cover, data.HeaderImg, data.Id)
	}, looklookTravelHomestayBusinessUserIdKey, looklookTravelHomestayBusinessIdKey)
	return err
}

func (m *defaultHomestayBusinessModel) Delete(session sqlx.Session, data *HomestayBusiness) error {
	data.DelState = globalkey.DelStateYes
	return m.Update(session, data)
}

func (m *defaultHomestayBusinessModel) Trans(fn func(session sqlx.Session) error) error {

	err := m.Transact(func(session sqlx.Session) error {
		return fn(session)
	})
	return err

}

func (m *defaultHomestayBusinessModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheLooklookTravelHomestayBusinessIdPrefix, primary)
}

func (m *defaultHomestayBusinessModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", homestayBusinessRows, m.table)
	return conn.QueryRow(v, query, primary)
}
