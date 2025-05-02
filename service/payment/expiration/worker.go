package expiration

import (
	"context"

	"dishes-service-backend/entity"
	"github.com/pkg/errors"
)

type OrderRepo interface {
	SetOrderStatus(ctx context.Context, orderId, oldStatus, newStatus string) error
}

type Worker struct {
	repo OrderRepo
}

func NewWorker(repo OrderRepo) Worker {
	return Worker{
		repo: repo,
	}
}

func (w Worker) ProcessPayment(ctx context.Context, req *PaymentPayload) error {
	err := w.repo.SetOrderStatus(ctx, req.OrderId, entity.OrderItemStatusProcess, entity.OrderItemStatusCanceled)
	if err != nil {
		return errors.WithMessage(err, "update order status")
	}
	return nil
}
