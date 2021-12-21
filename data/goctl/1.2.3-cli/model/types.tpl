
type (
	{{.upperStartCamelObject}}Model interface{
		{{.method}}
		Insert(session sqlx.Session,data *{{.upperStartCamelObject}}) (sql.Result,error)
		Update(session sqlx.Session,data *{{.upperStartCamelObject}}) error
		Delete(session sqlx.Session, data *{{.upperStartCamelObject}}) error
		Trans(fn func(session sqlx.Session) error) error
	}

	default{{.upperStartCamelObject}}Model struct {
		{{if .withCache}}sqlc.CachedConn{{else}}conn sqlx.SqlConn{{end}}
		table string
	}

	{{.upperStartCamelObject}} struct {
		{{.fields}}
	}
)
