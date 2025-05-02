package payment

import (
	"context"
	"github.com/Falokut/go-kit/log"
	"github.com/txix-open/bgjob"
	"time"
)

type Observer struct {
	log log.Logger
}

func NewObserver(log log.Logger) Observer {
	return Observer{
		log: log,
	}
}

func (o Observer) JobStarted(ctx context.Context, job bgjob.Job) {
	o.log.Debug(ctx, "bgjob: job started",
		log.Any("id", job.Id),
		log.Any("type", job.Type),
	)
}

func (o Observer) JobCompleted(ctx context.Context, job bgjob.Job) {
	o.log.Debug(ctx, "bgjob: job completed", log.Any("id", job.Id))
}

func (o Observer) JobWillBeRetried(ctx context.Context, job bgjob.Job, after time.Duration, err error) {
	o.log.Debug(ctx, "bgjob: job will be retried",
		log.Any("id", job.Id),
		log.Any("after", after.String()),
		log.Any("error", err),
	)
}

func (o Observer) JobMovedToDlq(ctx context.Context, job bgjob.Job, err error) {
	o.log.Debug(ctx, "bgjob: job will be moved to dlq",
		log.Any("id", job.Id),
		log.Any("error", err),
	)
}

func (o Observer) JobRescheduled(ctx context.Context, job bgjob.Job, after time.Duration) {
	o.log.Debug(ctx, "bgjob: job rescheduled",
		log.Any("id", job.Id),
		log.Any("after", after.String()),
	)
}

func (o Observer) QueueIsEmpty(ctx context.Context) {
}

func (o Observer) WorkerError(ctx context.Context, err error) {
	o.log.Debug(ctx, "bgjob: unexpected worker error", log.Any("error", err))
}
