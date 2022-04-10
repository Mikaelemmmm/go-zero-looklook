package logic

import (
	"context"
	"github.com/hibiken/asynq"
	"looklook/app/mqueue/cmd/job/internal/svc"
	"looklook/app/mqueue/cmd/job/jobtype"
)


type CronJob struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCronJob(ctx context.Context, svcCtx *svc.ServiceContext) *CronJob {
	return &CronJob{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// register job
func (l *CronJob) Register() *asynq.ServeMux {

	mux := asynq.NewServeMux()

	//scheduler job
	mux.Handle(jobtype.ScheduleSettleRecord,NewSettleRecordHandler(l.svcCtx))

	//defer job
	mux.Handle(jobtype.DeferCloseHomestayOrder,NewCloseHomestayOrderHandler(l.svcCtx))

	//queue job , asynq support queue job
	// wait you fill..

	return mux
}


