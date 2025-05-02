package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/Falokut/go-kit/http/apierrors"

	"dishes-service-backend/domain"
)

const (
	userIdHeader = "X-User-Id"
)

type OrderService interface {
	ProcessOrder(ctx context.Context, userId string, req domain.ProcessOrderRequest) (string, error)
	GetUserOrders(ctx context.Context, userId string, req domain.GetMyOrdersRequest) ([]domain.UserOrder, error)
}

type Order struct {
	service OrderService
}

func NewOrder(service OrderService) Order {
	return Order{
		service: service,
	}
}

// Process order
//
//	@Tags		order
//	@Summary	Заказать
//	@Accept		json
//	@Produce	json
//	@Security	Bearer
//	@Param		body	body		domain.ProcessOrderRequest	true	"request body"
//	@Success	200		{object}	domain.ProcessOrderResponse
//	@Failure	400		{object}	apierrors.Error
//	@Failure	404		{object}	apierrors.Error
//	@Failure	500		{object}	apierrors.Error
//	@Router		/orders [POST]
func (c Order) ProcessOrder(ctx context.Context, r *http.Request, req domain.ProcessOrderRequest) (*domain.ProcessOrderResponse, error) {
	url, err := c.service.ProcessOrder(ctx, r.Header.Get(userIdHeader), req)
	switch {
	case errors.Is(err, domain.ErrDishNotFound):
		return nil, apierrors.New(http.StatusNotFound, domain.ErrCodeDishNotFound, domain.ErrDishNotFound.Error(), err)
	case errors.Is(err, domain.ErrInvalidDishCount):
		return nil, apierrors.NewBusinessError(domain.ErrCodeInvalidDishCount, domain.ErrInvalidDishCount.Error(), err)
	case err != nil:
		return nil, err
	default:
		return &domain.ProcessOrderResponse{PaymentUrl: url}, nil
	}
}

// Get my orders
//
//	@Tags		order
//	@Summary	Получить заказы
//	@Produce	json
//	@Param		limit	query	int	false	"максимальное количество блюд"
//	@Param		offset	query	int	false	"смещение"
//	@Security	Bearer
//	@Success	200	{object}	domain.UserOrder
//	@Failure	400	{object}	apierrors.Error
//	@Failure	403	{object}	apierrors.Error
//	@Failure	404	{object}	apierrors.Error
//	@Failure	500	{object}	apierrors.Error
//	@Router		/orders/my [GET]
func (c Order) GetUserOrders(ctx context.Context,
	req domain.GetMyOrdersRequest,
	r *http.Request,
) ([]domain.UserOrder, error) {
	return c.service.GetUserOrders(ctx, r.Header.Get(userIdHeader), req)
}
