func (m *default{{.upperStartCamelObject}}Model) Update(session sqlx.Session,data *{{.upperStartCamelObject}}) error {
        {{if .withCache}}{{.keys}}
    	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
                query := fmt.Sprintf("update %s set %s where {{.originalPrimaryKey}} = {{if .postgreSql}}$1{{else}}?{{end}}", m.table, {{.lowerStartCamelObject}}RowsWithPlaceHolder)
				if session != nil{
					return session.Exec(query, {{.expressionValues}})	
				}
                return conn.Exec(query, {{.expressionValues}})
        }, {{.keyValues}}){{else}}query := fmt.Sprintf("update %s set %s where {{.originalPrimaryKey}} = {{if .postgreSql}}$1{{else}}?{{end}}", m.table, {{.lowerStartCamelObject}}RowsWithPlaceHolder)
    	var err error
		if session != nil{
			_,err=session.Exec(query, {{.expressionValues}})
		}else{
			_,err=m.conn.Exec(query, {{.expressionValues}})
		}
		{{end}}
        return err
}

