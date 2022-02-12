package request

import (
	"looklook/admin/model/common/request"
	"looklook/admin/model/system"
)

type SysDictionaryDetailSearch struct {
	system.SysDictionaryDetail
	request.PageInfo
}
