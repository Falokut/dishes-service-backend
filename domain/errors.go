package domain

import (
	"net/http"

	"github.com/Falokut/go-kit/http/apierrors"
	"github.com/pkg/errors"
)

var (
	ErrInvalidPaymentMethod       = errors.New("невалидный способ оплаты")
	ErrUserAlreadyExists          = errors.New("пользователь уже существует")
	ErrUserNotFound               = errors.New("пользователь не найден")
	ErrTelegramSignMissing        = errors.New("отсутствует подпись от telegram")
	ErrTelegramAuthDateMissing    = errors.New("отсутствует дата создания токена от telegram")
	ErrTelegramCredentialsExpired = errors.New("данные от telegram устарели")
	ErrInvalidTelegramCredentials = errors.New("невалидные данные от telegram")
	ErrUserOperationForbidden     = errors.New("данная операция запрещена для пользователя")
	ErrWrongSecret                = errors.New("неверный пароль")
	ErrDishNotFound               = errors.New("не все блюда были найдены")
	ErrInvalidDishCount           = errors.New("невалидное значение количества блюд")
	ErrDishCategoryNotFound       = errors.New("категория не найдена")
	ErrDishCategoryConflict       = errors.New("категория с таким именем уже существует")
	ErrRestaurantNotFound         = errors.New("ресторан не найдена")
	ErrRestaurantConflict         = errors.New("ресторан с таким названием уже существует")
	ErrInvalidToken               = errors.New("невалидный токен")
	ErrForbidden                  = errors.New("доступ запрещён")
	ErrOrderingForbidden          = errors.New("оформление заказов приостановлено")
	ErrOrderNotFound              = errors.New("заказ не найден")
)

const (
	ErrCodeInvalidArgument = 400

	ErrCodeInvalidDishCount     = 600
	ErrCodeDishNotFound         = 601
	ErrCodeDishCategoryNotFound = 602
	ErrCodeDishCategoryConflict = 603
	ErrCodeRestaurantNotFound   = 602
	ErrCodeRestaurantConflict   = 603
	ErrCodeUserNotFound         = 604
	ErrCodeUserAlreadyExists    = 605
	ErrCodeWrongSecret          = 606
	ErrCodeOrderingForbidden    = 607

	ErrCodeUnauthorized = 700
	ErrCodeForbidden    = 701
)

func DomainInvalidTokenError(err error) error {
	return apierrors.New(
		http.StatusUnauthorized,
		ErrCodeUnauthorized,
		ErrInvalidToken.Error(),
		err,
	)
}
