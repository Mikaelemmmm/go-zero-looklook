package kqueue

import (
	"encoding/json"
	"looklook/common/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-queue/kq"
)

type KqueueClient interface {
	Push(topic string, v interface{}) error
}

//封装的asynq业务客户端
type kqueueSvcClient struct {
	Brokers []string
}

func NewKqueueSvcClient(brokers []string) KqueueClient {
	return &kqueueSvcClient{
		Brokers: brokers,
	}
}

//添加
func (l *kqueueSvcClient) Push(topic string, v interface{}) error {

	//1、初始化pusher
	pusher := kq.NewPusher(l.Brokers, topic)

	//2、序列化
	body, err := json.Marshal(v)
	if err != nil {
		return errors.Wrapf(xerr.NewErrMsg("kq task marshal error "), "【pushKqMsgErrorMarshal】topic : %s , v : %+v", topic, v)
	}

	//3、发送消息.
	if err := pusher.Push(string(body)); err != nil {
		return errors.Wrapf(xerr.NewErrMsg("kq pusher task push error"), "【pushKqMsgErrorPush】topic : %s , v : %+v", topic, v)
	}

	return nil
}
