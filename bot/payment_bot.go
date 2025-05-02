package bot

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Falokut/go-kit/json"

	"dishes-service-backend/entity"

	"github.com/Falokut/go-kit/tg_bot"
	"github.com/pkg/errors"
)

type OrderService interface {
	UpdateOrderStatus(ctx context.Context, orderId string, newStatus string) error
	IsOrderingAllowed(ctx context.Context) (bool, error)
}

type BotAPI interface {
	Request(c tg_bot.Chattable) (*tg_bot.ApiResponse, error)
}
type PaymentBot struct {
	bot          BotAPI
	invoiceToken string
	service      OrderService
}

func NewPaymentBot(token string, bot BotAPI, service OrderService) PaymentBot {
	return PaymentBot{
		invoiceToken: token,
		bot:          bot,
		service:      service,
	}
}

const rubCurrency = "RUB"

func (b PaymentBot) ProcessPayment(ctx context.Context, order *entity.Order, chatId int64) error {
	isOrderingAllowed, err := b.service.IsOrderingAllowed(ctx)
	if err != nil {
		return errors.WithMessage(err, "get is ordering allowed")
	}
	if !isOrderingAllowed {
		return b.cancelOrder(ctx, order.Id)
	}
	payload := entity.PaymentPayload{
		ChatId:  chatId,
		OrderId: order.Id,
	}

	args, err := json.Marshal(payload)
	if err != nil {
		return errors.WithMessage(err, "marhal payload")
	}

	prices := make([]tg_bot.LabeledPrice, len(order.Items))
	for i := range order.Items {
		label := fmt.Sprintf("%s x %d", order.Items[i].Name, order.Items[i].Count)
		prices[i] = tg_bot.LabeledPrice{
			Label:  label,
			Amount: order.Items[i].Price,
		}
	}

	invoice := tg_bot.NewInvoice(
		chatId,
		fmt.Sprintf("Заказ № %s", order.Id),
		"оплата заказа",
		string(args),
		b.invoiceToken,
		"invoice",
		rubCurrency,
		prices,
	)
	resp, err := b.bot.Request(invoice)
	if err != nil {
		return errors.WithMessage(err, "send invoice")
	}
	switch {
	case resp.ErrorCode == http.StatusBadRequest:
		return b.cancelOrder(ctx, order.Id)
	case !resp.Ok:
		return errors.New("send invoice failed")
	}
	return nil
}

func (b PaymentBot) cancelOrder(ctx context.Context, orderId string) error {
	err := b.service.UpdateOrderStatus(ctx, orderId, entity.OrderItemStatusCanceled)
	if err != nil {
		return errors.WithMessage(err, "cancel order")
	}
	return nil
}
