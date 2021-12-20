package asynqmq

import (
	"encoding/json"
	"looklook/common/xerr"

	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
)

const (
	TypeHomestayOrderCloseDelivery = "homestay:order:close"
)

// 延迟关闭民宿订单task
type HomestayOrderCloseTaskPayload struct {
	Sn string
}

func NewHomestayOrderCloseTask(sn string) (*asynq.Task, error) {
	payload, err := json.Marshal(HomestayOrderCloseTaskPayload{Sn: sn})
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrMsg("创建延迟关闭民宿订单task到asynq失败"), "【addAsynqTaskMarshaError】err : %v , sn : %s", err, sn)
	}
	return asynq.NewTask(TypeHomestayOrderCloseDelivery, payload), nil
}
