package routes

import (
	"context"
	"dishes-service-backend/domain"
	"dishes-service-backend/entity"
	"net/http"
	"slices"

	http2 "github.com/Falokut/go-kit/http"
	"github.com/Falokut/go-kit/http/apierrors"
	"github.com/Falokut/go-kit/http/types"
	"github.com/Falokut/go-kit/jwt"
)

type AuthMiddleware struct {
	accessTokenSecret string
}

func NewAuthMiddleware(accessTokenSecret string) AuthMiddleware {
	return AuthMiddleware{
		accessTokenSecret: accessTokenSecret,
	}
}

func (m AuthMiddleware) AdminAuthToken() http2.Middleware {
	return AuthToken(m.accessTokenSecret, domain.AdminRoleName)
}

func (m AuthMiddleware) UserAuthToken() http2.Middleware {
	return AuthToken(m.accessTokenSecret, domain.UserRoleName, domain.AdminRoleName)
}

func AuthToken(tokenSecret string, roles ...string) http2.Middleware {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			token := &types.BearerToken{}
			err := token.FromRequestHeader(r)
			if err != nil {
				return err
			}

			userInfo := entity.TokenUserInfo{}
			err = jwt.ParseToken(token.Token, tokenSecret, &userInfo)
			if err != nil {
				return err
			}
			r.Header.Add(domain.UserIdHeader, userInfo.UserId)
			if len(roles) == 0 {
				return next(ctx, w, r)
			}
			if !slices.Contains(roles, userInfo.RoleName) {
				return apierrors.New(http.StatusForbidden,
					domain.ErrCodeForbidden,
					"доступ запрещён",
					domain.ErrForbidden, // nolint:err113
				)
			}
			return next(ctx, w, r)
		}
	}
}
