package wxMessage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"looklook/app/message/cmd/mq/internal/svc"
	"looklook/common/kqueue"
	"looklook/common/xerr"

	"github.com/pkg/errors"
	wechat "github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	miniConfig "github.com/silenceper/wechat/v2/miniprogram/config"
	"github.com/silenceper/wechat/v2/miniprogram/subscribe"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

/**
监听发送微信模版消息
*/
type SendWxMiniSubMessageMq struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendWxMiniSubMessageMq(ctx context.Context, svcCtx *svc.ServiceContext) *SendWxMiniSubMessageMq {
	return &SendWxMiniSubMessageMq{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendWxMiniSubMessageMq) Consume(_, val string) error {
	//解析数据.
	var message kqueue.SendWxMiniSubMessage
	if err := json.Unmarshal([]byte(val), &message); err != nil {
		logx.WithContext(l.ctx).Error("SendWxTplMq->Consume Unmarshal err : %v , val : %s", err, val)
		return err
	}
	//执行业务
	if err := l.execService(message); err != nil {
		logx.WithContext(l.ctx).Error("SendWxMiniSubMessageMq->execService  err : %v , val : %s , message:%+v", err, val, message)
		return err
	}
	return nil
}

//执行业务
func (l *SendWxMiniSubMessageMq) execService(message kqueue.SendWxMiniSubMessage) error {
	miniprogram := wechat.NewWechat().GetMiniProgram(&miniConfig.Config{
		AppID:     l.svcCtx.Config.WxMiniConf.AppId,
		AppSecret: l.svcCtx.Config.WxMiniConf.Secret,
		Cache:     cache.NewMemory(),
	})
	fmt.Printf("message :%+v \n\n", message)
	msg := &subscribe.Message{
		ToUser:     message.Openid,
		TemplateID: message.TemplateID,
		Page:       message.Page,
	}

	if len(message.Data) > 0 { //整合数据.
		//key2:val2#color2
		var msgData = make(map[string]*subscribe.DataItem, 2)
		for key, data := range message.Data {
			valColor := strings.Split(data, "#")
			var dataItem *subscribe.DataItem
			if len(valColor) == 2 {
				dataItem = &subscribe.DataItem{
					Value: valColor[0],
					Color: valColor[1],
				}
			} else {
				dataItem = &subscribe.DataItem{
					Value: valColor[0],
				}
			}

			fmt.Printf("key :%s ,dataItem :%+v \n", key, dataItem)
			msgData[key] = dataItem
		}

		msg.Data = msgData
	}

	//环境。
	if l.svcCtx.Config.Mode == service.DevMode {
		msg.MiniprogramState = "developer"
	} else {
		msg.MiniprogramState = "formal"
	}

	if err := miniprogram.GetSubscribe().Send(msg); err != nil {
		return errors.Wrapf(xerr.NewErrMsg("发送小程序订阅消息失败"), "发送小程序订阅消息失败 err:%v, msg ： %+v ", err, msg)
	}

	return nil
}
