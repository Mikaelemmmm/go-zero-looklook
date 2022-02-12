package request

import (
	"looklook/admin/model/common/request"
	"looklook/admin/model/system"
)

type SysOperationRecordSearch struct {
	system.SysOperationRecord
	request.PageInfo
}
