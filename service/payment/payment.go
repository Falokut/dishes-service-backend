package payment

import (
	"context"

	"dishes-service-backend/domain"
	"dishes-service-backend/entity"

	"github.com/pkg/errors"

	"github.com/Falokut/go-kit/log"
)

type PaymentService interface {
	Process(ctx context.Context, order *entity.Order) (string, error)
}

type ExpirationService interface {
	AddOrder(ctx context.Context, orderId string) error
}

type Payment struct {
	paymentMethods map[string]PaymentService
	expiration     ExpirationService
}

func NewPayment(
	logger log.Logger,
	paymentMethods map[string]PaymentService,
	expiration ExpirationService,
) Payment {
	return Payment{
		paymentMethods: paymentMethods,
		expiration:     expiration,
	}
}

func (s Payment) Process(ctx context.Context, order *entity.Order, method string) (string, error) {
	paymentService, ok := s.paymentMethods[method]
	if !ok {
		return "", domain.ErrInvalidPaymentMethod
	}
	err := s.expiration.AddOrder(ctx, order.Id)
	if err != nil {
		return "", errors.WithMessage(err, "add order to expiration")
	}

	url, err := paymentService.Process(ctx, order)
	if err != nil {
		return "", errors.WithMessage(err, "process payment")
	}

	return url, nil
}

func (s Payment) IsPaymentMethodValid(method string) bool {
	_, ok := s.paymentMethods[method]
	return ok
}
