
func (m *default{{.upperStartCamelObject}}Model) Delete(session sqlx.Session,data *{{.upperStartCamelObject}}) error {
	data.DelState = globalkey.DelStateYes
	return m.Update(session, data)
}



func (m *default{{.upperStartCamelObject}}Model) Trans(fn func(session sqlx.Session) error) error {
	{{if .withCache}}
		err := m.Transact(func(session sqlx.Session) error {
			return  fn(session)
		})
		return err
	{{else}}
		err := m.conn.Transact(func(session sqlx.Session) error {
			return  fn(session)
		})
		return err
	{{end}}
}