package payment

import (
	"context"

	"dishes-service-backend/entity"
	"github.com/pkg/errors"
)

type PaymentBot interface {
	ProcessPayment(ctx context.Context, order *entity.Order, chatId int64) error
}

type Worker struct {
	bot PaymentBot
}

func NewWorker(bot PaymentBot) Worker {
	return Worker{
		bot: bot,
	}
}

func (w Worker) ProcessPayment(ctx context.Context, req *PaymentPayload) error {
	err := w.bot.ProcessPayment(ctx, &req.Order, req.ChatId)
	if err != nil {
		return errors.WithMessage(err, "process payment")
	}
	return nil
}
