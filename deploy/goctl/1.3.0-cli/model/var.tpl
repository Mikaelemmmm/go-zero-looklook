
var (
	{{.lowerStartCamelObject}}FieldNames          = builder.RawFieldNames(&{{.upperStartCamelObject}}{}{{if .postgreSql}},true{{end}})
	{{.lowerStartCamelObject}}Rows                = strings.Join({{.lowerStartCamelObject}}FieldNames, ",")
	{{.lowerStartCamelObject}}RowsWithPlaceHolder = {{if .postgreSql}}builder.PostgreSqlJoin(stringx.Remove({{.lowerStartCamelObject}}FieldNames, "{{.originalPrimaryKey}}", "create_time", "update_time")){{else}}strings.Join(stringx.Remove({{.lowerStartCamelObject}}FieldNames, "{{.originalPrimaryKey}}", "`create_time`", "`update_time`", "`version`" ), "=?,") + "=?"{{end}}

	{{if .withCache}}{{.cacheKeys}}{{end}}
)
