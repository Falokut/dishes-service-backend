package events

import (
	"context"
	"dishes-service-backend/bot/routes"
	"github.com/Falokut/go-kit/tg_bot"
	"github.com/pkg/errors"
)

type Bot interface {
	Send(msg tg_bot.Chattable) error
}

type AdminEvents struct {
	tgBot Bot
}

func NewAdminEvents(tgBot Bot) AdminEvents {
	return AdminEvents{
		tgBot: tgBot,
	}
}

func (e AdminEvents) AdminAdded(ctx context.Context, chatId int64) error {
	scope := tg_bot.NewBotCommandScopeChat(chatId)
	endpoints := routes.EndpointsDescriptors(routes.Controllers{})
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
	err := e.tgBot.Send(tg_bot.NewSetMyCommandsWithScope(scope, commands...))
	if err != nil {
		return errors.WithMessage(err, "send admin commands")
	}
	return nil
}

func (e AdminEvents) AdminRemoved(ctx context.Context, chatId int64) error {
	scope := tg_bot.NewBotCommandScopeChat(chatId)
	err := e.tgBot.Send(tg_bot.NewDeleteMyCommandsWithScope(scope))
	if err != nil {
		return errors.WithMessage(err, "send delete chat scope commands")
	}
	return nil
}
