package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Falokut/go-kit/json"
	"github.com/pkg/errors"

	"dishes-service-backend/domain"
	"dishes-service-backend/entity"

	"github.com/Falokut/go-kit/tg_bot"
	"github.com/Falokut/go-kit/tg_botx/apierrors"
)

type OrderService interface {
	GetOrder(ctx context.Context, orderId string) (*entity.Order, error)
	SetOrderStatus(ctx context.Context, orderId string, oldStatus string, newStatus string) error
	GetOrderStatus(ctx context.Context, orderId string) (string, error)
	IsOrderingAllowed(ctx context.Context) (bool, error)
	SetOrderingAllowed(ctx context.Context, isAllowed bool) error
}

type OrderUserService interface {
	NotifySuccessPayment(ctx context.Context, req *entity.Order) error
	NotifyOrderArrival(ctx context.Context, req entity.QueryCallbackPayload) error
	CancelPaidOrder(ctx context.Context, req entity.QueryCallbackPayload) error
}

type CsvExporter interface {
	GetOrdersCsv(ctx context.Context, start time.Time, end time.Time) ([]byte, error)
}

type Order struct {
	orderService OrderService
	userService  OrderUserService
	cvsExporter  CsvExporter
}

func NewOrder(
	service OrderService,
	userService OrderUserService,
	cvsExporter CsvExporter,
) Order {
	return Order{
		orderService: service,
		userService:  userService,
		cvsExporter:  cvsExporter,
	}
}
func (c Order) HandlePayment(ctx context.Context, update tg_bot.Update) (tg_bot.Chattable, error) {
	msg := update.Message
	var payload entity.PaymentPayload
	err := json.Unmarshal([]byte(msg.SuccessfulPayment.InvoicePayload), &payload)
	if err != nil {
		return nil, apierrors.NewBusinessError(domain.ErrCodeInvalidArgument, "invalid payment payload", err)
	}

	err = c.orderService.SetOrderStatus(ctx, payload.OrderId, entity.OrderItemStatusProcess, entity.OrderItemStatusPaid)
	if err != nil {
		return nil, err
	}

	order, err := c.orderService.GetOrder(ctx, payload.OrderId)
	if err != nil {
		return nil, err
	}
	err = c.userService.NotifySuccessPayment(ctx, order)
	if err != nil {
		return nil, err
	}
	return nil, nil // nolint:nilnil
}

func (c Order) CsvOrdersInfo(ctx context.Context, update tg_bot.Update) (tg_bot.Chattable, error) {
	arguments := update.Message.CommandArguments()
	dates := strings.Split(arguments, "-")
	// nolint:mnd
	if len(dates) != 2 {
		return nil, apierrors.NewBusinessError(domain.ErrCodeInvalidArgument,
			"неправильный формат периода, должен быть: гггг.мм.дд-гггг.мм.дд",
			errors.New("invalid date format"),
		)
	}
	start, err := time.Parse(entity.DataFormat, dates[0])
	if err != nil {
		return nil, apierrors.NewBusinessError(domain.ErrCodeInvalidArgument,
			"неправильный формат начала периода, должен быть: гггг.мм.дд",
			err,
		)
	}
	end, err := time.Parse(entity.DataFormat, dates[1])
	if err != nil {
		return nil, apierrors.NewBusinessError(domain.ErrCodeInvalidArgument,
			"неправильный формат конца периода, должен быть: гггг.мм.дд",
			err,
		)
	}
	csvBody, err := c.cvsExporter.GetOrdersCsv(ctx, start, end)
	if err != nil {
		return nil, err
	}
	document := tg_bot.NewDocument(update.FromChat().Id, tg_bot.FileBytes{
		Name:  arguments + "_orders.csv",
		Bytes: csvBody,
	})
	document.Caption = fmt.Sprintf("отчёт по заказам с %s по %s", dates[0], dates[1])
	return document, nil
}

// nolint:nilerr
func (c Order) HandlePreCheckout(ctx context.Context, update tg_bot.Update) (tg_bot.Chattable, error) {
	query := update.PreCheckoutQuery
	var payload entity.PaymentPayload
	err := json.Unmarshal([]byte(query.InvoicePayload), &payload)
	if err != nil {
		return tg_bot.PreCheckoutConfig{
			PreCheckoutQueryID: query.Id,
			OK:                 false,
			ErrorMessage:       "invalid payload",
		}, nil
	}

	orderStatus, err := c.orderService.GetOrderStatus(ctx, payload.OrderId)
	if err != nil {
		return nil, apierrors.NewInternalServiceError(err)
	}
	switch {
	case orderStatus == entity.OrderItemStatusPaid:
		return tg_bot.PreCheckoutConfig{
			PreCheckoutQueryID: query.Id,
			OK:                 false,
			ErrorMessage:       "order already paid",
		}, nil
	case orderStatus == entity.OrderItemStatusCanceled:
		return tg_bot.PreCheckoutConfig{
			PreCheckoutQueryID: query.Id,
			OK:                 false,
			ErrorMessage:       "order canceled",
		}, nil
	}
	allowed, err := c.orderService.IsOrderingAllowed(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "is ordering allowed")
	}
	if !allowed {
		return tg_bot.PreCheckoutConfig{
			PreCheckoutQueryID: query.Id,
			OK:                 false,
			ErrorMessage:       domain.ErrOrderingForbidden.Error(),
		}, nil
	}

	return tg_bot.PreCheckoutConfig{
		PreCheckoutQueryID: query.Id,
		OK:                 true,
	}, nil
}
func (c Order) ForbidOrdering(ctx context.Context, update tg_bot.Update) (tg_bot.Chattable, error) {
	err := c.orderService.SetOrderingAllowed(ctx, false)
	if err != nil {
		return nil, err
	}
	return tg_bot.NewMessage(update.Message.Chat.Id, "оформление заказов запрещено"), nil
}
func (c Order) AllowOrdering(ctx context.Context, update tg_bot.Update) (tg_bot.Chattable, error) {
	err := c.orderService.SetOrderingAllowed(ctx, true)
	if err != nil {
		return nil, err
	}
	return tg_bot.NewMessage(update.Message.Chat.Id, "оформление заказов разрешено"), nil
}

func (c Order) HandleCallbackQuery(ctx context.Context, update tg_bot.Update) (tg_bot.Chattable, error) {
	var req entity.QueryCallbackPayload
	err := req.FromString(update.CallbackQuery.Data)
	if err != nil {
		return nil, apierrors.NewBusinessError(domain.ErrCodeInvalidArgument, "invalid callback query payload", err)
	}
	switch {
	case req.Command == entity.NotifyArrivalCommand:
		err = c.userService.NotifyOrderArrival(ctx, req)
		if err != nil {
			return nil, err
		}
	case req.Command == entity.SuccessOrderCommand:
		err = c.orderService.SetOrderStatus(ctx, req.OrderId, entity.OrderItemStatusPaid, entity.OrderItemStatusSuccess)
		if err != nil {
			return nil, err
		}
	case req.Command == entity.CancelOrderCommand:
		err = c.userService.CancelPaidOrder(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil //nolint:nilnil
}
