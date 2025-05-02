package payment

import (
	"context"
	"github.com/Falokut/go-kit/json"

	"dishes-service-backend/entity"
	"github.com/pkg/errors"
	"github.com/txix-open/bgjob"
)

type UserRepo interface {
	GetUserChatId(ctx context.Context, userId string) (int64, error)
}

type Payment struct {
	userRepo UserRepo
	cli      *bgjob.Client
}

func NewPayment(userRepo UserRepo, cli *bgjob.Client) Payment {
	return Payment{
		userRepo: userRepo,
		cli:      cli,
	}
}

const PaymentMethod string = "telegram"
const (
	WorkerQueue = "telegram-payment"
	WorkerType  = "payment"
)

func (s Payment) Process(ctx context.Context, order *entity.Order) (string, error) {
	chatId, err := s.userRepo.GetUserChatId(ctx, order.UserId)
	if err != nil {
		return "", errors.WithMessage(err, "get user chat id")
	}

	payload := PaymentPayload{
		Order:  *order,
		ChatId: chatId,
	}

	arg, err := json.Marshal(payload)
	if err != nil {
		return "", errors.WithMessage(err, "marshal payload")
	}

	err = s.cli.Enqueue(ctx, bgjob.EnqueueRequest{
		Id:    order.Id,
		Queue: WorkerQueue,
		Type:  WorkerType,
		Arg:   arg,
	},
	)

	if err != nil {
		return "", errors.WithMessage(err, "enqueue job")
	}
	return "", nil
}
