package routes

import (
	"dishes-service-backend/bot/controller"

	"github.com/Falokut/go-kit/tg_bot"
	"github.com/Falokut/go-kit/tg_botx/router"
)

type Controllers struct {
	User  controller.User
	Order controller.Order
}

type Endpoint struct {
	Handler     router.HandlerFunc
	Command     string
	Description string
	UpdateType  string
	Admin       bool
	Hide        bool
}

func InitRoutes(c Controllers, middlewares []router.Middleware, adminAuthMiddleware router.Middleware) *router.Router {
	router := router.NewRouter(middlewares...)
	for _, command := range EndpointsDescriptors(c) {
		if command.UpdateType == tg_bot.MessageUpdateType && command.Admin {
			router.Handler(command.UpdateType, command.Command, adminAuthMiddleware(command.Handler))
			continue
		}
		router.Handler(command.UpdateType, command.Command, command.Handler)
	}
	return router
}

func EndpointsDescriptors(c Controllers) []Endpoint {
	return []Endpoint{
		{
			Handler:     c.User.Register,
			Command:     "start",
			Description: "зарегистрироваться",
			UpdateType:  tg_bot.MessageUpdateType,
		},
		{
			Handler:     c.User.List,
			Command:     "user_list",
			Description: "получить список пользователей",
			UpdateType:  tg_bot.MessageUpdateType,
			Admin:       true,
		},
		{
			Command:     "add_admin",
			Handler:     c.User.AddAdmin,
			UpdateType:  tg_bot.MessageUpdateType,
			Description: "добавить админа по username",
			Admin:       true,
		},
		{
			Command:     "remove_admin",
			Handler:     c.User.RemoveAdminByUsername,
			UpdateType:  tg_bot.MessageUpdateType,
			Description: "удалить админа по username",
			Admin:       true,
		},
		{
			Command:    "pass_by_secret",
			Handler:    c.User.AddAdminSecret,
			UpdateType: tg_bot.MessageUpdateType,
			Hide:       true,
		},
		{
			Command:     "secret",
			Description: "Получить значение secret для становления админом",
			Handler:     c.User.GetAdminSecret,
			UpdateType:  tg_bot.MessageUpdateType,
			Admin:       true,
		},
		{
			Command:     "allow_ordering",
			Description: "Разрешить заказывать",
			Handler:     c.Order.AllowOrdering,
			UpdateType:  tg_bot.MessageUpdateType,
			Admin:       true,
		},
		{
			Command:     "forbid_ordering",
			Description: "Запретить заказывать",
			Handler:     c.Order.ForbidOrdering,
			UpdateType:  tg_bot.MessageUpdateType,
			Admin:       true,
		},
		{
			Handler:     c.Order.CsvOrdersInfo,
			UpdateType:  tg_bot.MessageUpdateType,
			Command:     "order_csv",
			Description: "получить csv файл с информацией о заказах в указанный период гггг.мм.дд-гггг.мм.дд",
			Admin:       true,
		},
		{
			Handler:    c.Order.HandleCallbackQuery,
			UpdateType: tg_bot.CallbackQueryUpdateType,
		},
		{
			Handler:    c.Order.HandlePreCheckout,
			UpdateType: tg_bot.PreCheckoutQueryUpdateType,
		},
		{
			Handler:    c.Order.HandlePayment,
			UpdateType: tg_bot.SuccessfulPaymentMessageUpdateType,
			Hide:       true,
		},
	}
}
