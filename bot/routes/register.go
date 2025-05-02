package routes

import (
	"context"
	"dishes-service-backend/entity"

	"github.com/Falokut/go-kit/tg_bot"
	"github.com/Falokut/go-kit/tg_botx"
	"github.com/pkg/errors"
)

type AdminUsersProvides interface {
	GetTelegramUsersInfo(ctx context.Context) ([]entity.TelegramUser, error)
}

func RegisterRoutes(ctx context.Context, tgBot *tg_botx.Bot, usersProvider AdminUsersProvides) error {
	users, err := usersProvider.GetTelegramUsersInfo(ctx)
	if err != nil {
		return errors.WithMessage(err, "get telegram users info")
	}

	err = tgBot.ClearAllCommands()
	if err != nil {
		return errors.WithMessage(err, "clear commands")
	}
	commands := defaultCommands()
	err = tgBot.RegisterCommands(commands...)
	if err != nil {
		return errors.WithMessage(err, "register default commands")
	}

	scopes, commands := adminCommands(users)
	for _, scope := range scopes {
		err = tgBot.Send(tg_bot.NewSetMyCommandsWithScope(scope, commands...))
		if err != nil {
			return errors.WithMessage(err, "register admin commands")
		}
	}

	return nil
}

// func clearCommandForChat(tgBot *tg_bot.BotApi, chatId int64) error {
// 	scope := tg_bot.NewBotCommandScopeChat(chatId)
// 	err := tgBot.Send(tg_bot.NewDeleteMyCommandsWithScope(scope))
// 	if err != nil {
// 		return errors.WithMessage(err, "send delete chat scope commands")
// 	}
// 	return nil
// }

// func clearCommands(tgBot *tg_bot.BotApi, users []entity.TelegramUser) error {
// 	err := tgBot.Send(tg_bot.NewDeleteMyCommands())
// 	if err != nil {
// 		return errors.WithMessage(err, "send delete default scope commands")
// 	}
// 	for _, user := range users {
// 		err = clearCommandForChat(tgBot, user.ChatId)
// 		if err != nil {
// 			return errors.WithMessagef(err, "clear commands for chatId=%d", user.ChatId)
// 		}
// 	}
// 	return nil
// }

func defaultCommands() []tg_bot.BotCommand {
	endpoints := EndpointsDescriptors(Controllers{})
	commands := make([]tg_bot.BotCommand, 0, len(endpoints))
	for _, endpoint := range endpoints {
		if endpoint.Hide || endpoint.Admin || endpoint.UpdateType != tg_bot.MessageUpdateType {
			continue
		}
		commands = append(commands, tg_bot.BotCommand{
			Command:     endpoint.Command,
			Description: endpoint.Description,
		})
	}
	return commands
}

func adminCommands(users []entity.TelegramUser) ([]tg_bot.BotCommandScope, []tg_bot.BotCommand) {
	endpoints := EndpointsDescriptors(Controllers{})
	commands := make([]tg_bot.BotCommand, 0, len(endpoints))
	for _, endpoint := range endpoints {
		if endpoint.Hide || endpoint.UpdateType != tg_bot.MessageUpdateType {
			continue
		}
		commands = append(commands, tg_bot.BotCommand{
			Command:     endpoint.Command,
			Description: endpoint.Description,
		})
	}
	scopes := make([]tg_bot.BotCommandScope, 0)
	for _, user := range users {
		if !user.Admin {
			continue
		}
		scopes = append(scopes, tg_bot.NewBotCommandScopeChat(user.ChatId))
	}
	return scopes, commands
}
