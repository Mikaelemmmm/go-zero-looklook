
func (m *default{{.upperStartCamelObject}}Model) Insert(session sqlx.Session, data *{{.upperStartCamelObject}}) (sql.Result,error) {
	{{if .withCache}}{{if .containsIndexCache}}{{.keys}}
    return m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {

		//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
		query := fmt.Sprintf("insert into .... (%s) values ...", m.table)
		if session != nil{
			//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
			return session.Exec(query,{{.expressionValues}})	
		}
		//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
		return conn.Exec(query, {{.expressionValues}})
	}, {{.keyValues}}){{else}}
	//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
	query := fmt.Sprintf("insert into .... (%s) values ...", m.table)
	if session != nil{
		//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
	   return session.Exec(query, {{.expressionValues}})	
	}
	//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
    return m.ExecNoCache(query, {{.expressionValues}})
	{{end}}{{else}}
	//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
	query := fmt.Sprintf("insert into .... (%s) values ...", m.table)
	if session != nil{
		//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
		return session.Exec(query,{{.expressionValues}})	
	}
	//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
	return m.conn.Exec(query, {{.expressionValues}})
	
	{{end}}
	
}
