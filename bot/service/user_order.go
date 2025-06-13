package service

import (
	"context"
	"dishes-service-backend/entity"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/Falokut/go-kit/tg_bot"
	"github.com/pkg/errors"
)

type UserRepo interface {
	GetUserInfo(ctx context.Context, userId string) (entity.User, error)
	GetAdminsChatsIds(ctx context.Context) ([]int64, error)
	GetUserChatId(ctx context.Context, userId string) (int64, error)
}

type OrderRepo interface {
	GetOrderedChatId(ctx context.Context, orderId string) (int64, error)
	SetOrderStatus(ctx context.Context, orderId, oldStatus, newStatus string) error
}
type BotAPI interface {
	Send(c tg_bot.Chattable) error
}

type UserOrder struct {
	bot       BotAPI
	userRepo  UserRepo
	orderRepo OrderRepo
}

func NewOrderUserService(bot BotAPI, userRepo UserRepo, orderRepo OrderRepo) UserOrder {
	return UserOrder{
		bot:       bot,
		userRepo:  userRepo,
		orderRepo: orderRepo,
	}
}

func (s UserOrder) NotifySuccessPayment(ctx context.Context, order *entity.Order) error {
	adminIds, err := s.userRepo.GetAdminsChatsIds(ctx)
	if err != nil {
		return errors.WithMessage(err, "get admins chats ids")
	}
	user, err := s.userRepo.GetUserInfo(ctx, order.UserId)
	if err != nil {
		return errors.WithMessage(err, "get user info")
	}

	orderInfoString := s.getOrderInfoString(order, &user)
	markup := s.getMarkupForOrder(order)
	for _, chatId := range adminIds {
		message := tg_bot.NewMessage(chatId, orderInfoString)
		message.ReplyMarkup = markup
		message.ParseMode = tg_bot.ModeHTML
		err = s.bot.Send(message)
		if err != nil {
			return errors.WithMessagef(err, "send notification to chat: %d", chatId)
		}
	}
	return nil
}

func (s UserOrder) getMarkupForOrder(order *entity.Order) tg_bot.InlineKeyboardMarkup {
	arrivalPayload := entity.QueryCallbackPayload{
		Command: entity.NotifyArrivalCommand,
		OrderId: order.Id,
	}
	notifyArrivalButton := tg_bot.NewInlineKeyboardButtonData(
		"оповестить о прибытии заказа",
		arrivalPayload.String(),
	)
	cancelPayload := entity.QueryCallbackPayload{
		Command: entity.CancelOrderCommand,
		OrderId: order.Id,
	}
	cancelButton := tg_bot.NewInlineKeyboardButtonData(
		"отменить заказ",
		cancelPayload.String(),
	)

	markup := tg_bot.NewInlineKeyboardMarkup(
		[]tg_bot.InlineKeyboardButton{notifyArrivalButton},
		[]tg_bot.InlineKeyboardButton{cancelButton},
	)
	return markup
}

func (s UserOrder) NotifyOrderArrival(ctx context.Context, req entity.QueryCallbackPayload) error {
	chatId, err := s.orderRepo.GetOrderedChatId(ctx, req.OrderId)
	if err != nil {
		return errors.WithMessage(err, "get user chat id")
	}

	button := tg_bot.NewInlineKeyboardButtonData("подтвердить получение",
		entity.QueryCallbackPayload{Command: entity.SuccessOrderCommand, OrderId: req.OrderId}.String(),
	)
	message := tg_bot.NewMessage(chatId, fmt.Sprintf("Заказ №%s прибыл", req.OrderId))
	message.ReplyMarkup = tg_bot.NewInlineKeyboardMarkup([]tg_bot.InlineKeyboardButton{button})
	err = s.bot.Send(message)
	if err != nil {
		return errors.WithMessagef(err, "send notification to chat: %d", chatId)
	}
	return nil
}

func (s UserOrder) CancelPaidOrder(ctx context.Context, req entity.QueryCallbackPayload) error {
	err := s.orderRepo.SetOrderStatus(ctx, req.OrderId, entity.OrderItemStatusPaid, entity.OrderItemStatusCanceled)
	if err != nil {
		return errors.WithMessage(err, "update order status")
	}
	chatId, err := s.orderRepo.GetOrderedChatId(ctx, req.OrderId)
	if err != nil {
		return errors.WithMessage(err, "get user chat id")
	}
	err = s.bot.Send(tg_bot.NewMessage(chatId, fmt.Sprintf("Заказ №%s отменён", req.OrderId)))
	if err != nil {
		return errors.WithMessagef(err, "send notification to chat: %d", chatId)
	}
	return nil
}

// nolint:gosmopolitan,mnd
func (s UserOrder) getOrderInfoString(order *entity.Order, user *entity.User) string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "<b>Заказ №%s</b>\n\n", html.EscapeString(order.Id))
	builder.WriteString("<b>Состав заказа:</b>\n")

	itemsByRestaurant := make(map[string][]entity.OrderItem, len(order.Items))
	for _, item := range order.Items {
		itemsByRestaurant[item.RestaurantName] = append(itemsByRestaurant[item.RestaurantName], item)
	}

	for restName, items := range itemsByRestaurant {
		builder.WriteString(fmt.Sprintf("<u>Ресторан: %s</u>\n", html.EscapeString(restName)))
		builder.WriteString("<code>ID   Название                     Кол-во</code>\n")
		builder.WriteString("<code>---  --------------------------  ------</code>\n")
		for _, item := range items {
			name := html.EscapeString(item.Name)
			if len(name) > 24 {
				name = name[:21] + "..."
			}
			line := fmt.Sprintf("<code>%-4d %-26s %6d</code>\n", item.DishId, name, item.Count)
			builder.WriteString(line)
		}
		builder.WriteString("\n")
	}

	fmt.Fprintf(&builder, "<b>Имя заказавшего:</b> %s\n", html.EscapeString(user.Name))
	fmt.Fprintf(&builder, "<b>Telegram ник:</b> @%s\n", html.EscapeString(user.Username))
	fmt.Fprintf(&builder, "<b>Стоимость:</b> %d.%02d руб\n", order.Total/100, order.Total%100)
	fmt.Fprintf(&builder, "<b>Пожелания:</b> '%s'\n", html.EscapeString(order.Wishes))
	fmt.Fprintf(&builder, "<b>Дата:</b> %s", html.EscapeString(order.CreatedAt.Local().Format(time.DateTime)))

	return builder.String()
}
