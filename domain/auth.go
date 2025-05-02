package domain

import (
	"github.com/Falokut/go-kit/jwt"
)

const (
	AuthHeaderName = "Authorization"
	UserIdHeader   = "X-User-Id"
	BearerToken    = "Bearer"
)

const (
	AdminRoleName = "admin"
	UserRoleName  = "user"
)

type LoginByTelegramRequest struct {
	InitTelegramData string
}

type LoginResponse struct {
	AccessToken  jwt.TokenResponse
	RefreshToken jwt.TokenResponse
}

type UserRoleResponse struct {
	RoleName string
}
