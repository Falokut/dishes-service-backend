package controller

import (
	"context"
	"errors"
	"net/http"

	"dishes-service-backend/domain"

	"github.com/Falokut/go-kit/http/apierrors"
	_ "github.com/Falokut/go-kit/http/apierrors"
)

type DishService interface {
	List(ctx context.Context, limit, offset int32) ([]domain.Dish, error)
	GetByIds(ctx context.Context, ids []int32) ([]domain.Dish, error)
	GetByCategories(ctx context.Context, limit, offset int32, ids []int32) ([]domain.Dish, error)
	AddDish(ctx context.Context, req domain.AddDishRequest) (*domain.AddDishResponse, error)
	EditDish(ctx context.Context, req domain.EditDishRequest) error
	DeleteDish(ctx context.Context, id int32) error
}

const (
	maxGetDishesCount = 30
)

type Dish struct {
	service DishService
}

func NewDish(service DishService) Dish {
	return Dish{
		service: service,
	}
}

// List
//
//	@Tags			dishes
//	@Summary		dish
//	@Description	возвращает список блюд
//	@Param			ids			query	string	false	"список идентификаторов блюд через запятую"
//	@Param			сategories	query	string	false	"список идентификаторов категорий через запятую"
//	@Param			limit		query	int		false	"максимальное количество блюд"
//	@Param			offset		query	int		false	"смещение"
//	@Produce		json
//	@Success		200	{array}		domain.Dish
//	@Failure		400	{object}	apierrors.Error
//	@Failure		500	{object}	apierrors.Error
//	@Router			/dishes [GET]
func (c Dish) List(ctx context.Context, req domain.GetDishesRequest) ([]domain.Dish, error) {
	ids, err := stringToIntSlice(req.Ids)
	if err != nil {
		return nil, apierrors.NewBusinessError(domain.ErrCodeInvalidArgument, "invalid ids", err)
	}
	if len(ids) > maxGetDishesCount {
		return nil, apierrors.NewBusinessError(domain.ErrCodeInvalidArgument, "invalid ids count", err)
	}
	categoriesIds, err := stringToIntSlice(req.CategoriesIds)
	if err != nil {
		return nil, apierrors.NewBusinessError(domain.ErrCodeInvalidArgument, "invalid categories ids", err)
	}
	switch {
	case len(ids) > 0:
		return c.service.GetByIds(ctx, ids)
	case len(categoriesIds) > 0:
		return c.service.GetByCategories(ctx, req.Limit, req.Offset, categoriesIds)
	default:
		return c.service.List(ctx, req.Limit, req.Offset)
	}
}

// Add dish
//
//	@Tags		dishes
//	@Summary	Add Dish
//	@Param		body	body	domain.AddDishRequest	true	"request body"
//
//	@Security	Bearer
//
//	@Accept		json
//	@Success	200	{object}	domain.AddDishResponse
//	@Failure	403	{object}	apierrors.Error
//	@Failure	500	{object}	apierrors.Error
//	@Router		/dishes [POST]
func (c Dish) AddDish(ctx context.Context, req domain.AddDishRequest) (*domain.AddDishResponse, error) {
	return c.service.AddDish(ctx, req)
}

// Edit dish
//
//	@Tags		dishes
//	@Summary	Edit Dish
//	@Param		body	body	domain.EditDishRequest	true	"request body"
//	@Param		id		path	int32					true	"идентификатор блюда"
//
//	@Security	Bearer
//
//	@Accept		json
//	@Success	200	{object}	any
//	@Failure	400	{object}	apierrors.Error
//	@Failure	403	{object}	apierrors.Error
//	@Failure	404	{object}	apierrors.Error
//	@Failure	500	{object}	apierrors.Error
//	@Router		/dishes/edit/{id} [POST]
func (c Dish) EditDish(ctx context.Context, req domain.EditDishRequest) error {
	err := c.service.EditDish(ctx, req)
	switch {
	case errors.Is(err, domain.ErrDishNotFound):
		return apierrors.New(http.StatusNotFound, domain.ErrCodeDishNotFound, domain.ErrDishNotFound.Error(), err)
	default:
		return err
	}
}

// Delete dish
//
//	@Tags		dishes
//	@Summary	Delete Dish
//
//	@Param		id	path	int32	true	"идентификатор блюда"
//
//	@Security	Bearer
//
//	@Success	200	{object}	any
//	@Failure	400	{object}	apierrors.Error
//	@Failure	403	{object}	apierrors.Error
//	@Failure	404	{object}	apierrors.Error
//	@Failure	500	{object}	apierrors.Error
//	@Router		/dishes/delete/{id} [DELETE]
func (c Dish) DeleteDish(ctx context.Context, req domain.DeleteDishRequest) error {
	err := c.service.DeleteDish(ctx, req.Id)
	switch {
	case errors.Is(err, domain.ErrDishNotFound):
		return apierrors.New(http.StatusNotFound, domain.ErrCodeDishNotFound, domain.ErrDishNotFound.Error(), err)
	default:
		return err
	}
}
