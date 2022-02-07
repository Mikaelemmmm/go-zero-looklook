package logic

import (
	"context"

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

	if err := l.svcCtx.KqueueClient.Push(kqueue.SEND_WX_MINI_TPL_MESSAGE, m); err != nil {
		return nil, err
	}

	return &pb.SendWxMiniSubMessageResp{}, nil
}
