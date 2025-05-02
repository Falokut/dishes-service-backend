package service

import (
	"context"

	"dishes-service-backend/domain"
	"dishes-service-backend/entity"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type UserRepo interface {
	IsAdmin(ctx context.Context, id string) (bool, error)
	GetUsers(ctx context.Context) ([]entity.User, error)
	GetUserChatId(ctx context.Context, userId string) (int64, error)
	GetUserIdByTelegramId(ctx context.Context, chatId int64) (string, error)
	SetUserAdminStatus(ctx context.Context, username string, isAdmin bool) error
	AddAdminChatId(ctx context.Context, chatId int64) error
	GetAdminsIds(ctx context.Context) ([]string, error)
	GetUserChatIdByUsername(ctx context.Context, username string) (int64, error)
}

type SecretRepo interface {
	GetSecret() string
}

type RegisterTx interface {
	InsertUser(ctx context.Context, user entity.InsertUser) (string, error)
	InsertUserTelegram(ctx context.Context, userId string, tg entity.Telegram) error
}

type UserTxRunner interface {
	RegisterUserTx(ctx context.Context, tx func(ctx context.Context, tx RegisterTx) error) error
}

type AdminEvents interface {
	AdminAdded(ctx context.Context, chatId int64) error
	AdminRemoved(ctx context.Context, chatId int64) error
}

type User struct {
	userRepo    UserRepo
	txRunner    UserTxRunner
	secretRepo  SecretRepo
	adminEvents AdminEvents
}

func NewUser(
	userRepo UserRepo,
	txRunner UserTxRunner,
	secretRepo SecretRepo,
	adminEvents AdminEvents,
) User {
	return User{
		userRepo:    userRepo,
		txRunner:    txRunner,
		secretRepo:  secretRepo,
		adminEvents: adminEvents,
	}
}

func (s User) Register(ctx context.Context, req domain.RegisterUser) error {
	err := s.txRunner.RegisterUserTx(ctx, func(ctx context.Context, tx RegisterTx) error {
		err := s.register(ctx, req, tx)
		if err != nil {
			return errors.WithMessage(err, "register")
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "register user tx")
	}
	return nil
}

func (s User) register(ctx context.Context, req domain.RegisterUser, tx RegisterTx) error {
	user := entity.InsertUser{
		Id:       uuid.NewString(),
		Name:     req.Name,
		Username: req.Username,
	}
	userId, err := tx.InsertUser(ctx, user)
	if err != nil {
		return errors.WithMessage(err, "insert user")
	}
	if req.Telegram == nil {
		return nil
	}

	tg := entity.Telegram{
		ChatId: req.Telegram.ChatId,
		UserId: req.Telegram.UserId,
	}
	err = tx.InsertUserTelegram(ctx, userId, tg)
	if err != nil {
		return errors.WithMessage(err, "insert user telegram")
	}
	return nil
}

func (s User) IsAdmin(ctx context.Context, userId string) (bool, error) {
	isAdmin, err := s.userRepo.IsAdmin(ctx, userId)
	if err != nil {
		return false, errors.WithMessage(err, "check is user admin")
	}
	return isAdmin, nil
}

func (s User) List(ctx context.Context) ([]domain.User, error) {
	users, err := s.userRepo.GetUsers(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get users")
	}

	converted := make([]domain.User, len(users))
	for i, u := range users {
		converted[i] = domain.User{
			Username: u.Username,
			Name:     u.Name,
			Admin:    u.Admin,
		}
	}
	return converted, nil
}

func (s User) AddAdmin(ctx context.Context, username string) error {
	err := s.userRepo.SetUserAdminStatus(ctx, username, true)
	if err != nil {
		return errors.WithMessage(err, "add admin")
	}
	chatId, err := s.userRepo.GetUserChatIdByUsername(ctx, username)
	if err != nil {
		return errors.WithMessage(err, "get chat id by username")
	}
	err = s.adminEvents.AdminAdded(ctx, chatId)
	if err != nil {
		return errors.WithMessage(err, "admin added event")
	}
	return nil
}

func (s User) GetChatId(ctx context.Context, userId string) (int64, error) {
	chatId, err := s.userRepo.GetUserChatId(ctx, userId)
	if err != nil {
		return -1, errors.WithMessage(err, "add admin")
	}
	return chatId, nil
}

func (s User) GetUserIdByTelegramId(ctx context.Context, telegramId int64) (string, error) {
	userId, err := s.userRepo.GetUserIdByTelegramId(ctx, telegramId)
	if err != nil {
		return "", errors.WithMessage(err, "get user id by telegram id")
	}
	return userId, nil
}

func (s User) AddAdminSecret(ctx context.Context, req domain.AddAdminSecretRequest) error {
	if s.secretRepo.GetSecret() != req.Secret {
		return domain.ErrWrongSecret
	}

	err := s.userRepo.AddAdminChatId(ctx, req.ChatId)
	if err != nil {
		return errors.WithMessage(err, "add admin")
	}
	err = s.adminEvents.AdminAdded(ctx, req.ChatId)
	if err != nil {
		return errors.WithMessage(err, "admin added event")
	}
	return nil
}

func (s User) GetAdminSecret(_ context.Context) (string, error) {
	return s.secretRepo.GetSecret(), nil
}

func (s User) RemoveAdmin(ctx context.Context, username string) error {
	err := s.userRepo.SetUserAdminStatus(ctx, username, false)
	if err != nil {
		return errors.WithMessage(err, "remove admin")
	}

	chatId, err := s.userRepo.GetUserChatIdByUsername(ctx, username)
	if err != nil {
		return errors.WithMessage(err, "get user chat id by username")
	}

	err = s.adminEvents.AdminRemoved(ctx, chatId)
	if err != nil {
		return errors.WithMessage(err, "admin removed")
	}
	return nil
}
