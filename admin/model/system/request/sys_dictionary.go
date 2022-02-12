package request

import (
	"looklook/admin/model/common/request"
	"looklook/admin/model/system"
)

type SysDictionarySearch struct {
	system.SysDictionary
	request.PageInfo
}
