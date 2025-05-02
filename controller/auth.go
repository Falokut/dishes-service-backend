package controller

import (
	"context"
	"dishes-service-backend/domain"
	"net/http"

	"github.com/Falokut/go-kit/http/apierrors"
	"github.com/Falokut/go-kit/http/types"
	"github.com/Falokut/go-kit/jwt"
	"github.com/pkg/errors"
)

type AuthService interface {
	LoginByTelegram(ctx context.Context, req domain.LoginByTelegramRequest) (*domain.LoginResponse, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*jwt.TokenResponse, error)
	GetUserRole(ctx context.Context, accessToken string) (*domain.UserRoleResponse, error)
}

type Auth struct {
	service AuthService
}

func NewAuth(service AuthService) Auth {
	return Auth{
		service: service,
	}
}

// LoginByTelegram
//
//	@Tags		auth
//	@Summary	Войти в аккаунт
//	@Accept		json
//	@Produce	json
//
//	@Param		body	body		domain.LoginByTelegramRequest	true	"тело запроса"
//
//	@Success	200		{object}	domain.LoginResponse
//	@Failure	404		{object}	apierrors.Error
//	@Failure	500		{object}	apierrors.Error
//	@Router		/auth/login_by_telegram [POST]
func (c Auth) LoginByTelegram(ctx context.Context, req domain.LoginByTelegramRequest) (*domain.LoginResponse, error) {
	tokens, err := c.service.LoginByTelegram(ctx, req)
	switch {
	case errors.Is(err, domain.ErrUserNotFound), errors.Is(err, domain.ErrInvalidTelegramCredentials):
		return nil, apierrors.New(
			http.StatusNotFound,
			domain.ErrCodeForbidden,
			domain.ErrForbidden.Error(),
			err,
		)
	default:
		return tokens, err
	}
}

// RefreshAccessToken
//
//	@Tags		auth
//	@Summary	Обновить токен доступа
//	@Accept		json
//	@Produce	json
//
//
//	@Success	200	{object}	jwt.TokenResponse
//	@Failure	404	{object}	apierrors.Error
//	@Failure	401	{object}	apierrors.Error
//	@Failure	500	{object}	apierrors.Error
//
//	@Security	Bearer
//
//	@Router		/auth/access_token [GET]
func (c Auth) RefreshAccessToken(ctx context.Context, token types.BearerToken) (*jwt.TokenResponse, error) {
	return c.service.RefreshAccessToken(ctx, token.Token)
}

// GetUserRole
//
//	@Tags		auth
//	@Summary	Получить роль пользователя
//	@Accept		json
//	@Produce	json
//
//
//	@Success	200	{object}	domain.UserRoleResponse
//	@Failure	404	{object}	apierrors.Error
//	@Failure	401	{object}	apierrors.Error
//	@Failure	500	{object}	apierrors.Error
//
//	@Security	Bearer
//
//	@Router		/auth/user_role [GET]
func (c Auth) GetUserRole(ctx context.Context, token types.BearerToken) (*domain.UserRoleResponse, error) {
	return c.service.GetUserRole(ctx, token.Token)
}
