package repository

import (
	"context"
	"database/sql"

	"dishes-service-backend/domain"
	"dishes-service-backend/entity"

	"github.com/Falokut/go-kit/db"
	"github.com/pkg/errors"
)

type User struct {
	cli db.DB
}

func NewUser(cli db.DB) User {
	return User{
		cli: cli,
	}
}

func (r User) InsertUser(ctx context.Context, user entity.InsertUser) (string, error) {
	query := `INSERT INTO users (id, username, name) VALUES($1,$2,$3) ON CONFLICT DO NOTHING RETURNING id`
	var userId string
	err := r.cli.SelectRow(ctx, &userId, query, user.Id, user.Username, user.Name)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", domain.ErrUserAlreadyExists
	case err != nil:
		return "", errors.WithMessagef(err, "exec query '%s'", query)
	}
	return userId, nil
}

func (r User) InsertUserTelegram(ctx context.Context, userId string, tg entity.Telegram) error {
	query := `INSERT INTO users_telegrams (id, chat_id, telegram_id) VALUES($1,$2,$3) ON CONFLICT DO NOTHING`
	_, err := r.cli.Exec(ctx, query, userId, tg.ChatId, tg.UserId)
	if err != nil {
		return errors.WithMessagef(err, "exec query '%s'", query)
	}
	return nil
}

func (r User) GetUserInfo(ctx context.Context, userId string) (entity.User, error) {
	query := "SELECT username,name,admin FROM users WHERE id=$1"
	var user entity.User
	err := r.cli.SelectRow(ctx, &user, query, userId)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return entity.User{}, domain.ErrUserNotFound
	case err != nil:
		return entity.User{}, errors.WithMessagef(err, "exec query '%s'", query)
	default:
		return user, nil
	}
}

func (r User) IsAdmin(ctx context.Context, id string) (bool, error) {
	query := `SELECT admin FROM users WHERE id=$1`

	var isAdmin bool
	err := r.cli.SelectRow(ctx, &isAdmin, query, id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return false, domain.ErrUserNotFound
	case err != nil:
		return false, errors.WithMessagef(err, "exec query '%s'", query)
	default:
		return isAdmin, nil
	}
}

func (r User) GetUsers(ctx context.Context) ([]entity.User, error) {
	query := "SELECT username, name, admin FROM users"
	var res []entity.User
	err := r.cli.Select(ctx, &res, query)
	if err != nil {
		return nil, errors.WithMessagef(err, "exec query '%s'", query)
	}

	return res, nil
}

func (r User) GetUserChatId(ctx context.Context, userId string) (int64, error) {
	query := "SELECT chat_id FROM users_telegrams WHERE id=$1"
	var chatId int64
	err := r.cli.SelectRow(ctx, &chatId, query, userId)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return 0, domain.ErrUserNotFound
	case err != nil:
		return 0, errors.WithMessagef(err, "exec query '%s'", query)
	default:
		return chatId, nil
	}
}

func (r User) GetUserByTelegramId(ctx context.Context, telegramId int64) (entity.User, error) {
	query := `
	SELECT u.id, u.username, u.name, u.admin
	FROM users u
	JOIN users_telegrams ut ON u.id=ut.id
	WHERE ut.telegram_id=$1`
	var user entity.User
	err := r.cli.SelectRow(ctx, &user, query, telegramId)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return entity.User{}, domain.ErrUserNotFound
	case err != nil:
		return entity.User{}, errors.WithMessagef(err, "exec query '%s'", query)
	default:
		return user, nil
	}
}

func (r User) SetUserAdminStatus(ctx context.Context, username string, isAdmin bool) error {
	query := "UPDATE users SET admin=$1 WHERE username=$2"
	res, err := r.cli.Exec(ctx, query, isAdmin, username)
	if err != nil {
		return errors.WithMessagef(err, "exec query '%s'", query)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r User) AddAdminChatId(ctx context.Context, chatId int64) error {
	query := `
	UPDATE users
	SET admin='true'
	FROM users_telegrams t
	WHERE users.id=t.id AND t.chat_id=$1`
	res, err := r.cli.Exec(ctx, query, chatId)
	if err != nil {
		return errors.WithMessagef(err, "exec query '%s'", query)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r User) GetAdminsIds(ctx context.Context) ([]string, error) {
	var ids []string
	err := r.cli.Select(ctx, &ids, "SELECT id FROM users WHERE admin='true'")
	if err != nil {
		return nil, errors.WithMessage(err, "select admins ids")
	}
	return ids, nil
}

func (r User) GetAdminsChatsIds(ctx context.Context) ([]int64, error) {
	query := "SELECT chat_id FROM users_telegrams t JOIN users u ON t.id=u.id WHERE u.admin"
	var chatIds []int64
	err := r.cli.Select(ctx, &chatIds, query)
	if err != nil {
		return nil, errors.WithMessagef(err, "exec query '%s'", query)
	}
	return chatIds, nil
}

func (r User) GetTelegramUsersInfo(ctx context.Context) ([]entity.TelegramUser, error) {
	query := `SELECT chat_id, admin 
	FROM users_telegrams t
	JOIN users u
	ON t.id=u.id;`
	var telegrams []entity.TelegramUser
	err := r.cli.Select(ctx, &telegrams, query)
	if err != nil {
		return nil, errors.WithMessagef(err, "exec query '%s'", query)
	}
	return telegrams, nil
}

func (r User) GetUserChatIdByUsername(ctx context.Context, username string) (int64, error) {
	query := `
	SELECT chat_id 
	FROM users_telegrams t
	JOIN users u
	ON t.id=u.id
	WHERE u.username=$1;
	`
	var chatId int64
	err := r.cli.SelectRow(ctx, &chatId, query, username)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return 0, domain.ErrUserNotFound
	case err != nil:
		return 0, errors.WithMessagef(err, "exec query '%s'", query)
	}
	return chatId, nil
}

func (r User) GetUserIdByTelegramId(ctx context.Context, telegramId int64) (string, error) {
	query := "SELECT id FROM users_telegrams WHERE telegram_id=$1"
	var userId string
	err := r.cli.SelectRow(ctx, &userId, query, telegramId)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", domain.ErrUserNotFound
	case err != nil:
		return "", errors.WithMessagef(err, "exec query '%s'", query)
	default:
		return userId, nil
	}
}
