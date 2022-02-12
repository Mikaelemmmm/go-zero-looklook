package request

import (
	"looklook/admin/model/autocode"
	"looklook/admin/model/common/request"
)

type {{.StructName}}Search struct{
    autocode.{{.StructName}}
    request.PageInfo
}
