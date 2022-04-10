package logic

import (
	"context"
	"fmt"
	"github.com/hibiken/asynq"
	"looklook/app/mqueue/cmd/job/internal/svc"
)


// SettleRecordHandler   shcedule billing to home business
type SettleRecordHandler struct {
	svcCtx *svc.ServiceContext
}

func NewSettleRecordHandler(svcCtx *svc.ServiceContext) *SettleRecordHandler {
	return &SettleRecordHandler{
		svcCtx:svcCtx,
	}
}

//  every one minute exec : if return err != nil , asynq will retry
func (l *SettleRecordHandler) ProcessTask(ctx context.Context, _ *asynq.Task) error {

	fmt.Printf("shcedule job demo -----> every one minute exec \n")

	return nil
}


