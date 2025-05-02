package service

import (
	"context"
	"dishes-service-backend/domain"
	"dishes-service-backend/entity"

	"github.com/pkg/errors"
)

type DishesRestaurantsRepo interface {
	GetAllRestaurants(ctx context.Context) ([]entity.Restaurant, error)
	GetRestaurant(ctx context.Context, id int32) (entity.Restaurant, error)
	InsertRestaurant(ctx context.Context, restaurant string) (int32, error)
	RenameRestaurant(ctx context.Context, id int32, newName string) error
	DeleteRestaurant(ctx context.Context, id int32) error
}

type Restaurant struct {
	repo DishesRestaurantsRepo
}

func NewRestaurant(repo DishesRestaurantsRepo) Restaurant {
	return Restaurant{repo: repo}
}

func (s Restaurant) GetAllRestaurants(ctx context.Context) ([]domain.Restaurant, error) {
	restaurants, err := s.repo.GetAllRestaurants(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get all restaurants")
	}
	domainRestaurants := make([]domain.Restaurant, len(restaurants))
	for i, restaurant := range restaurants {
		domainRestaurants[i] = domain.Restaurant{
			Id:   restaurant.Id,
			Name: restaurant.Name,
		}
	}
	return domainRestaurants, nil
}

func (s Restaurant) GetRestaurant(ctx context.Context, id int32) (domain.Restaurant, error) {
	restaurant, err := s.repo.GetRestaurant(ctx, id)
	if err != nil {
		return domain.Restaurant{}, errors.WithMessage(err, "get restaurant")
	}
	return domain.Restaurant{
		Id:   restaurant.Id,
		Name: restaurant.Name,
	}, nil
}

func (s Restaurant) AddRestaurant(ctx context.Context, restaurant string) (int32, error) {
	id, err := s.repo.InsertRestaurant(ctx, restaurant)
	if err != nil {
		return 0, errors.WithMessage(err, "add restaurant")
	}
	return id, nil
}

func (s Restaurant) RenameRestaurant(ctx context.Context, req domain.RenameRestaurantRequest) error {
	err := s.repo.RenameRestaurant(ctx, req.Id, req.Name)
	if err != nil {
		return errors.WithMessage(err, "rename restaurant")
	}
	return nil
}

func (s Restaurant) DeleteRestaurant(ctx context.Context, id int32) error {
	err := s.repo.DeleteRestaurant(ctx, id)
	if err != nil {
		return errors.WithMessage(err, "delete restaurant")
	}
	return nil
}
