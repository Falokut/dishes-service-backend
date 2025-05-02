package expiration

import (
	"context"
	"github.com/Falokut/go-kit/json"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/bgjob"
)

type Expiration struct {
	cli             *bgjob.Client
	expirationDelay time.Duration
}

func NewExpiration(cli *bgjob.Client, expirationDelay time.Duration) Expiration {
	return Expiration{
		cli:             cli,
		expirationDelay: expirationDelay,
	}
}

const PaymentMethod string = "telegram"
const (
	WorkerQueue = "payment-expiration"
	WorkerType  = "payment"
)

func (s Expiration) AddOrder(ctx context.Context, orderId string) error {
	payload := PaymentPayload{
		OrderId: orderId,
	}

	arg, err := json.Marshal(payload)
	if err != nil {
		return errors.WithMessage(err, "marshal payload")
	}

	err = s.cli.Enqueue(ctx, bgjob.EnqueueRequest{
		Queue: WorkerQueue,
		Type:  WorkerType,
		Arg:   arg,
		Delay: s.expirationDelay,
	},
	)

	if err != nil {
		return errors.WithMessage(err, "enqueue job")
	}
	return nil
}
