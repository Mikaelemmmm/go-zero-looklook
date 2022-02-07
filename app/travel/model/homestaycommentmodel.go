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
	homestayCommentFieldNames          = builder.RawFieldNames(&HomestayComment{})
	homestayCommentRows                = strings.Join(homestayCommentFieldNames, ",")
	homestayCommentRowsWithPlaceHolder = strings.Join(stringx.Remove(homestayCommentFieldNames, "`id`", "`create_time`", "`update_time`", "`version`"), "=?,") + "=?"

	cacheLooklookTravelHomestayCommentIdPrefix = "cache:looklookTravel:homestayComment:id:"
)

type (
	HomestayCommentModel interface {
		FindOne(id int64) (*HomestayComment, error)
		Insert(session sqlx.Session, data *HomestayComment) (sql.Result, error)
		Update(session sqlx.Session, data *HomestayComment) error
		Delete(session sqlx.Session, data *HomestayComment) error
		Trans(fn func(session sqlx.Session) error) error
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
	}
)

func NewHomestayCommentModel(conn sqlx.SqlConn, c cache.CacheConf) HomestayCommentModel {
	return &defaultHomestayCommentModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "`homestay_comment`",
	}
}

func (m *defaultHomestayCommentModel) Insert(session sqlx.Session, data *HomestayComment) (sql.Result, error) {

	//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
	query := fmt.Sprintf("insert into .... (%s) values ...", m.table)
	if session != nil {
		//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
		return session.Exec(query, data.DeleteTime, data.DelState, data.HomestayId, data.UserId, data.Content, data.Star)
	}
	//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
	return m.ExecNoCache(query, data.DeleteTime, data.DelState, data.HomestayId, data.UserId, data.Content, data.Star)

}

func (m *defaultHomestayCommentModel) FindOne(id int64) (*HomestayComment, error) {
	looklookTravelHomestayCommentIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayCommentIdPrefix, id)
	var resp HomestayComment
	err := m.QueryRow(&resp, looklookTravelHomestayCommentIdKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", homestayCommentRows, m.table)
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

func (m *defaultHomestayCommentModel) Update(session sqlx.Session, data *HomestayComment) error {
	looklookTravelHomestayCommentIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayCommentIdPrefix, data.Id)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, homestayCommentRowsWithPlaceHolder)
		if session != nil {
			return session.Exec(query, data.DeleteTime, data.DelState, data.HomestayId, data.UserId, data.Content, data.Star, data.Id)
		}
		return conn.Exec(query, data.DeleteTime, data.DelState, data.HomestayId, data.UserId, data.Content, data.Star, data.Id)
	}, looklookTravelHomestayCommentIdKey)
	return err
}

func (m *defaultHomestayCommentModel) Delete(session sqlx.Session, data *HomestayComment) error {
	data.DelState = globalkey.DelStateYes
	return m.Update(session, data)
}

func (m *defaultHomestayCommentModel) Trans(fn func(session sqlx.Session) error) error {

	err := m.Transact(func(session sqlx.Session) error {
		return fn(session)
	})
	return err

}

func (m *defaultHomestayCommentModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheLooklookTravelHomestayCommentIdPrefix, primary)
}

func (m *defaultHomestayCommentModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", homestayCommentRows, m.table)
	return conn.QueryRow(v, query, primary)
}
