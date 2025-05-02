package controller

import (
	"context"
	"dishes-service-backend/domain"
	"errors"
	"net/http"

	"github.com/Falokut/go-kit/http/apierrors"
)

type RestaurantService interface {
	GetAllRestaurants(ctx context.Context) ([]domain.Restaurant, error)
	GetRestaurant(ctx context.Context, id int32) (domain.Restaurant, error)
	AddRestaurant(ctx context.Context, category string) (int32, error)
	RenameRestaurant(ctx context.Context, req domain.RenameRestaurantRequest) error
	DeleteRestaurant(ctx context.Context, id int32) error
}
type Restaurant struct {
	service RestaurantService
}

func NewRestaurant(service RestaurantService) Restaurant {
	return Restaurant{service: service}
}

// Get all restaurants
//
//	@Tags		restaurants
//	@Summary	Получить все рестораны
//	@Accept		json
//	@Produce	json
//	@Success	200	{array}		domain.Restaurant
//	@Failure	500	{object}	apierrors.Error
//	@Router		/restaurants [GET]
func (c Restaurant) GetAllRestaurants(ctx context.Context) ([]domain.Restaurant, error) {
	return c.service.GetAllRestaurants(ctx)
}

// Get category
//
//	@Tags		restaurants
//	@Summary	Получить ресторан
//	@Produce	json
//	@Param		body	body		domain.GetDishesRestaurant	true	"request body"
//	@Param		id		path		int32						true	"Идентификатор ресторана"
//
//	@Success	200		{object}	domain.Restaurant
//	@Failure	400		{object}	apierrors.Error
//	@Failure	404		{object}	apierrors.Error
//	@Failure	500		{object}	apierrors.Error
//	@Router		/restaurants/{id} [GET]
func (c Restaurant) GetRestaurant(ctx context.Context, req domain.GetDishesRestaurant) (*domain.Restaurant, error) {
	category, err := c.service.GetRestaurant(ctx, req.Id)
	switch {
	case errors.Is(err, domain.ErrRestaurantNotFound):
		return nil, apierrors.New(http.StatusNotFound,
			domain.ErrCodeRestaurantNotFound,
			domain.ErrRestaurantNotFound.Error(),
			err,
		)
	case err != nil:
		return nil, err
	default:
		return &category, nil
	}
}

// Add category
//
//	@Param		body	body	domain.AddRestaurantRequest	true	"request body"
//
//	@Security	Bearer
//
//	@Tags		restaurants
//	@Summary	Создать ресторан
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	domain.Restaurant
//	@Failure	400	{string}	apierrors.Error
//	@Failure	403	{object}	apierrors.Error
//	@Failure	500	{object}	apierrors.Error
//	@Router		/restaurants [POST]
func (c Restaurant) AddRestaurant(ctx context.Context, req domain.AddRestaurantRequest) (*domain.AddRestaurantResponse, error) {
	id, err := c.service.AddRestaurant(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	return &domain.AddRestaurantResponse{Id: id}, nil
}

// Rename category
//
//	@Param		body	body	domain.RenameRestaurantRequest	true	"request body"
//
//	@Security	Bearer

// @Tags		restaurants
// @Summary	Переименовать ресторан
// @Accept		json
// @Produce	json
//
// @Param		id		path		int								true	"Идентификатор ресторана"
// @Param		body	body		domain.RenameRestaurantRequest	true	"request body"
//
// @Success	204		{object}	any
// @Failure	400		{object}	apierrors.Error
// @Failure	403		{object}	apierrors.Error
// @Failure	500		{object}	apierrors.Error
// @Router		/restaurants/{id} [POST]
func (c Restaurant) RenameRestaurant(ctx context.Context, req domain.RenameRestaurantRequest) error {
	err := c.service.RenameRestaurant(ctx, req)
	switch {
	case errors.Is(err, domain.ErrRestaurantConflict):
		return apierrors.New(
			http.StatusConflict,
			domain.ErrCodeRestaurantConflict,
			domain.ErrRestaurantConflict.Error(),
			err,
		)
	case err != nil:
		return err
	default:
		return nil
	}
}

// Delete category
//
//	@Tags		restaurants
//	@Summary	Удалить ресторан
//	@Produce	json
//
//	@Param		body	body	domain.DeleteRestaurantRequest	true	"request body"
//	@Param		id		path	int32							true	"Идентификатор ресторана"
//
//	@Security	Bearer
//
//	@Success	204	{object}	any
//	@Failure	400	{object}	apierrors.Error
//	@Failure	500	{object}	apierrors.Error
//	@Router		/restaurants/{id} [DELETE]
func (c Restaurant) DeleteRestaurant(ctx context.Context, req domain.DeleteRestaurantRequest) error {
	return c.service.DeleteRestaurant(ctx, req.Id)
}
