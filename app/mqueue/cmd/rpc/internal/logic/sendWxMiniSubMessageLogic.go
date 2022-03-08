package logic

import (
	"context"
	"github.com/pkg/errors"
	"looklook/common/xerr"
	"encoding/json"

	"looklook/app/mqueue/cmd/rpc/internal/svc"
	"looklook/app/mqueue/cmd/rpc/pb"
	"looklook/common/kqueue"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendWxMiniSubMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendWxMiniSubMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendWxMiniSubMessageLogic {
	return &SendWxMiniSubMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发送微信小程序订阅消息
func (l *SendWxMiniSubMessageLogic) SendWxMiniSubMessage(in *pb.SendWxMiniSubMessageReq) (*pb.SendWxMiniSubMessageResp, error) {

	m := kqueue.SendWxMiniSubMessage{
		Openid:     in.Openid,
		Data:       in.Data,
		TemplateID: in.TemplateID,
		Page:       in.Page,
	}

	//2、序列化
	body, err := json.Marshal(m)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrMsg("kq sendWxMiniSubMessageLogic task marshal error "), "kq sendWxMiniSubMessageLogic task marshal error , v : %+v", m)
	}

	if err := l.svcCtx.KqueueSendWxMiniTplMessageClient.Push(string(body)); err != nil {
		return nil, err
	}

	return &pb.SendWxMiniSubMessageResp{}, nil
}
