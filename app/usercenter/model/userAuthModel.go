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
	userAuthFieldNames          = builder.RawFieldNames(&UserAuth{})
	userAuthRows                = strings.Join(userAuthFieldNames, ",")
	userAuthRowsWithPlaceHolder = strings.Join(stringx.Remove(userAuthFieldNames, "`id`", "`create_time`", "`update_time`", "`version`"), "=?,") + "=?"

	cacheLooklookUsercenterUserAuthIdPrefix              = "cache:looklookUsercenter:userAuth:id:"
	cacheLooklookUsercenterUserAuthAuthTypeAuthKeyPrefix = "cache:looklookUsercenter:userAuth:authType:authKey:"
	cacheLooklookUsercenterUserAuthUserIdAuthTypePrefix  = "cache:looklookUsercenter:userAuth:userId:authType:"
)

type (
	UserAuthModel interface {
		FindOne(id int64) (*UserAuth, error)
		FindOneByAuthTypeAuthKey(authType string, authKey string) (*UserAuth, error)
		FindOneByUserIdAuthType(userId int64, authType string) (*UserAuth, error)
		Insert(session sqlx.Session, data *UserAuth) (sql.Result, error)
		Update(session sqlx.Session, data *UserAuth) error
		Delete(session sqlx.Session, data *UserAuth) error
		Trans(fn func(session sqlx.Session) error) error
	}

	defaultUserAuthModel struct {
		sqlc.CachedConn
		table string
	}

	UserAuth struct {
		Id         int64     `db:"id"`
		CreateTime time.Time `db:"create_time"`
		UpdateTime time.Time `db:"update_time"`
		DeleteTime time.Time `db:"delete_time"`
		DelState   int64     `db:"del_state"`
		UserId     int64     `db:"user_id"`
		AuthKey    string    `db:"auth_key"`  // 平台唯一id
		AuthType   string    `db:"auth_type"` // 平台类型
	}
)

func NewUserAuthModel(conn sqlx.SqlConn, c cache.CacheConf) UserAuthModel {
	return &defaultUserAuthModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "`user_auth`",
	}
}

func (m *defaultUserAuthModel) Insert(session sqlx.Session, data *UserAuth) (sql.Result, error) {
	looklookUsercenterUserAuthIdKey := fmt.Sprintf("%s%v", cacheLooklookUsercenterUserAuthIdPrefix, data.Id)
	looklookUsercenterUserAuthAuthTypeAuthKeyKey := fmt.Sprintf("%s%v:%v", cacheLooklookUsercenterUserAuthAuthTypeAuthKeyPrefix, data.AuthType, data.AuthKey)
	looklookUsercenterUserAuthUserIdAuthTypeKey := fmt.Sprintf("%s%v:%v", cacheLooklookUsercenterUserAuthUserIdAuthTypePrefix, data.UserId, data.AuthType)
	return m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {

		query := fmt.Sprintf("insert into %s(user_id,auth_key,auth_type) values (?,?,?)", m.table)
		if session != nil {
			return session.Exec(query, data.UserId, data.AuthKey, data.AuthType)
		}
		return conn.Exec(query, data.UserId, data.AuthKey, data.AuthType)
	}, looklookUsercenterUserAuthIdKey, looklookUsercenterUserAuthAuthTypeAuthKeyKey, looklookUsercenterUserAuthUserIdAuthTypeKey)

}

func (m *defaultUserAuthModel) FindOne(id int64) (*UserAuth, error) {
	looklookUsercenterUserAuthIdKey := fmt.Sprintf("%s%v", cacheLooklookUsercenterUserAuthIdPrefix, id)
	var resp UserAuth
	err := m.QueryRow(&resp, looklookUsercenterUserAuthIdKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", userAuthRows, m.table)
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

func (m *defaultUserAuthModel) FindOneByAuthTypeAuthKey(authType string, authKey string) (*UserAuth, error) {
	looklookUsercenterUserAuthAuthTypeAuthKeyKey := fmt.Sprintf("%s%v:%v", cacheLooklookUsercenterUserAuthAuthTypeAuthKeyPrefix, authType, authKey)
	var resp UserAuth
	err := m.QueryRowIndex(&resp, looklookUsercenterUserAuthAuthTypeAuthKeyKey, m.formatPrimary, func(conn sqlx.SqlConn, v interface{}) (i interface{}, e error) {
		query := fmt.Sprintf("select %s from %s where `auth_type` = ? and `auth_key` = ? limit 1", userAuthRows, m.table)
		if err := conn.QueryRow(&resp, query, authType, authKey); err != nil {
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

func (m *defaultUserAuthModel) FindOneByUserIdAuthType(userId int64, authType string) (*UserAuth, error) {
	looklookUsercenterUserAuthUserIdAuthTypeKey := fmt.Sprintf("%s%v:%v", cacheLooklookUsercenterUserAuthUserIdAuthTypePrefix, userId, authType)
	var resp UserAuth
	err := m.QueryRowIndex(&resp, looklookUsercenterUserAuthUserIdAuthTypeKey, m.formatPrimary, func(conn sqlx.SqlConn, v interface{}) (i interface{}, e error) {
		query := fmt.Sprintf("select %s from %s where `user_id` = ? and `auth_type` = ? limit 1", userAuthRows, m.table)
		if err := conn.QueryRow(&resp, query, userId, authType); err != nil {
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

func (m *defaultUserAuthModel) Update(session sqlx.Session, data *UserAuth) error {
	looklookUsercenterUserAuthIdKey := fmt.Sprintf("%s%v", cacheLooklookUsercenterUserAuthIdPrefix, data.Id)
	looklookUsercenterUserAuthAuthTypeAuthKeyKey := fmt.Sprintf("%s%v:%v", cacheLooklookUsercenterUserAuthAuthTypeAuthKeyPrefix, data.AuthType, data.AuthKey)
	looklookUsercenterUserAuthUserIdAuthTypeKey := fmt.Sprintf("%s%v:%v", cacheLooklookUsercenterUserAuthUserIdAuthTypePrefix, data.UserId, data.AuthType)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, userAuthRowsWithPlaceHolder)
		if session != nil {
			return session.Exec(query, data.DeleteTime, data.DelState, data.UserId, data.AuthKey, data.AuthType, data.Id)
		}
		return conn.Exec(query, data.DeleteTime, data.DelState, data.UserId, data.AuthKey, data.AuthType, data.Id)
	}, looklookUsercenterUserAuthAuthTypeAuthKeyKey, looklookUsercenterUserAuthUserIdAuthTypeKey, looklookUsercenterUserAuthIdKey)
	return err
}

func (m *defaultUserAuthModel) Delete(session sqlx.Session, data *UserAuth) error {
	data.DelState = globalkey.DelStateYes
	return m.Update(session, data)
}

func (m *defaultUserAuthModel) Trans(fn func(session sqlx.Session) error) error {

	err := m.Transact(func(session sqlx.Session) error {
		return fn(session)
	})
	return err

}

func (m *defaultUserAuthModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheLooklookUsercenterUserAuthIdPrefix, primary)
}

func (m *defaultUserAuthModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", userAuthRows, m.table)
	return conn.QueryRow(v, query, primary)
}
