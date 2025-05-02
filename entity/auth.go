package entity

import (
	"errors"
	"fmt"
)

const (
	userIdJwtKey   = "user_id"
	roleNameJwtKey = "role"
)

type TokenUserInfo struct {
	UserId   string `json:"user_id"`
	RoleName string `json:"role"`
}

func (t *TokenUserInfo) ToMap() map[string]any {
	return map[string]any{
		userIdJwtKey:   t.UserId,
		roleNameJwtKey: t.RoleName,
	}
}
func (t *TokenUserInfo) FromMap(m map[string]any) error {
	userId, ok := m[userIdJwtKey]
	if !ok {
		return errors.New("user id key not found")
	}
	roleName, ok := m[roleNameJwtKey]
	if !ok {
		return errors.New("role name key not found")
	}

	t.UserId = fmt.Sprint(userId)
	t.RoleName = fmt.Sprint(roleName)
	return nil
}

type UserAuthInfo struct {
	UserId   string
	RoleName string
}
