package expiration

import (
	"context"
	"github.com/Falokut/go-kit/json"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/bgjob"
)

type PaymentWorker interface {
	ProcessPayment(ctx context.Context, req *PaymentPayload) error
}

type WorkerController struct {
	worker PaymentWorker
}

func NewWorkerController(worker PaymentWorker) WorkerController {
	return WorkerController{
		worker: worker,
	}
}

const defaultRetryTime = time.Minute * 5

//nolint:gocritic
func (c WorkerController) Handle(ctx context.Context, job bgjob.Job) bgjob.Result {
	var payload PaymentPayload
	err := json.Unmarshal(job.Arg, &payload)
	if err != nil {
		return bgjob.MoveToDlq(errors.WithMessage(err, "unmarshal payload"))
	}

	err = c.worker.ProcessPayment(ctx, &payload)
	if err != nil {
		return bgjob.Reschedule(defaultRetryTime)
	}

	return bgjob.Complete()
}
