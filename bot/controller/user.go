package controller

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"dishes-service-backend/domain"

	"github.com/Falokut/go-kit/tg_bot"
	"github.com/Falokut/go-kit/tg_botx/apierrors"
)

type UserService interface {
	Register(ctx context.Context, user domain.RegisterUser) error
	IsAdmin(ctx context.Context, userId string) (bool, error)
	GetUserIdByTelegramId(ctx context.Context, telegramId int64) (string, error)
	List(ctx context.Context) ([]domain.User, error)
	AddAdmin(ctx context.Context, username string) error
	RemoveAdmin(ctx context.Context, username string) error
	AddAdminSecret(ctx context.Context, req domain.AddAdminSecretRequest) error
	GetAdminSecret(ctx context.Context) (string, error)
}

type User struct {
	service UserService
}

func NewUser(service UserService) User {
	return User{
		service: service,
	}
}

func (c User) Register(ctx context.Context, update tg_bot.Update) (tg_bot.Chattable, error) {
	msg := update.Message
	name := msg.From.FirstName
	if msg.From.LastName != "" {
		name += " " + msg.From.LastName
	}
	user := domain.RegisterUser{
		Username: msg.From.UserName,
		Name:     name,
		Telegram: &domain.Telegram{
			ChatId: msg.Chat.Id,
			UserId: msg.From.Id,
		},
	}
	err := c.service.Register(ctx, user)
	switch {
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return nil, apierrors.NewBusinessError(domain.ErrCodeUserAlreadyExists, domain.ErrUserAlreadyExists.Error(), err)
	case err != nil:
		return nil, err
	}
	return tg_bot.NewMessage(msg.Chat.Id, "вы зарегистрированы"), nil
}

const userUnderline = "___________________________________________"

//nolint:mnd
func (c User) List(ctx context.Context, update tg_bot.Update) (tg_bot.Chattable, error) {
	users, err := c.service.List(ctx)
	if err != nil {
		return nil, err
	}
	text := make([]string, 0, len(users)*2+3)
	text = append(text, userUnderline, "|  #  |  [NAME]  |  [USERNAME]  |  [ADMIN]  |")
	for i, user := range users {
		text = append(text, userUnderline, fmt.Sprintf("|  %d  |%s|%s|%t|",
			i+1, user.Name, user.Username, user.Admin))
	}
	return tg_bot.NewMessage(update.Message.Chat.Id, strings.Join(text, "\n")), nil
}

func (c User) AddAdmin(ctx context.Context, update tg_bot.Update) (tg_bot.Chattable, error) {
	msg := update.Message
	err := c.service.AddAdmin(ctx, msg.CommandArguments())
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return nil, apierrors.NewBusinessError(
			domain.ErrCodeUserNotFound,
			fmt.Sprintf("пользователь с ником '%s' не найден", msg.CommandArguments()),
			err)
	case err != nil:
		return nil, err
	}
	return tg_bot.NewMessage(msg.Chat.Id,
			fmt.Sprintf("администратор с username '%s' добавлен", msg.CommandArguments()),
		),
		nil
}

func (c User) RemoveAdminByUsername(ctx context.Context, update tg_bot.Update) (tg_bot.Chattable, error) {
	msg := update.Message
	err := c.service.RemoveAdmin(ctx, msg.CommandArguments())
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return nil, apierrors.NewBusinessError(
			domain.ErrCodeUserNotFound,
			fmt.Sprintf("пользователь с username '%s' не найден", msg.CommandArguments()),
			err)
	case err != nil:
		return nil, err
	}
	return tg_bot.NewMessage(msg.Chat.Id,
			fmt.Sprintf("администратор с username '%s' удалён", msg.CommandArguments()),
		),
		nil
}

func (c User) AddAdminSecret(ctx context.Context, update tg_bot.Update) (tg_bot.Chattable, error) {
	msg := update.Message
	req := domain.AddAdminSecretRequest{
		ChatId: msg.Chat.Id,
		Secret: msg.CommandArguments(),
	}
	err := c.service.AddAdminSecret(ctx, req)
	switch {
	case err != nil:
		return nil, err
	case errors.Is(err, domain.ErrWrongSecret):
		return nil, apierrors.NewBusinessError(domain.ErrCodeWrongSecret, domain.ErrWrongSecret.Error(), err)
	case errors.Is(err, domain.ErrUserNotFound):
		return nil, apierrors.NewBusinessError(domain.ErrCodeUserNotFound,
			fmt.Sprintf("пользователь с username '%s' не найден", msg.From.UserName),
			err,
		)
	}
	return tg_bot.NewMessage(msg.Chat.Id, "вы стали администратором"), nil
}

func (c User) GetAdminSecret(ctx context.Context, update tg_bot.Update) (tg_bot.Chattable, error) {
	secret, err := c.service.GetAdminSecret(ctx)
	if err != nil {
		return nil, err
	}
	return tg_bot.NewMessage(update.Message.Chat.Id, "пароль для администратора: "+secret), nil
}
