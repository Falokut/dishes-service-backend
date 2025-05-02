package service

import (
	"context"
	"dishes-service-backend/domain"
	"dishes-service-backend/entity"

	"github.com/pkg/errors"
)

type DishCategoryRepo interface {
	GetAllCategories(ctx context.Context) ([]entity.DishCategory, error)
	GetDishCategory(ctx context.Context) ([]entity.DishCategory, error)
	GetCategory(ctx context.Context, id int32) (entity.DishCategory, error)
	AddCategory(ctx context.Context, category string) (int32, error)
	RenameCategory(ctx context.Context, id int32, newName string) error
	DeleteCategory(ctx context.Context, id int32) error
}

type DishCategory struct {
	repo DishCategoryRepo
}

func NewDishCategory(repo DishCategoryRepo) DishCategory {
	return DishCategory{repo: repo}
}

func (s DishCategory) GetAllCategories(ctx context.Context) ([]domain.DishCategory, error) {
	categories, err := s.repo.GetAllCategories(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get all categories")
	}
	domainCategories := make([]domain.DishCategory, len(categories))
	for i, category := range categories {
		domainCategories[i] = domain.DishCategory{
			Id:   category.Id,
			Name: category.Name,
		}
	}
	return domainCategories, nil
}

func (s DishCategory) GetDishCategory(ctx context.Context) ([]domain.DishCategory, error) {
	categories, err := s.repo.GetDishCategory(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get dishes categories")
	}
	domainCategories := make([]domain.DishCategory, len(categories))
	for i, category := range categories {
		domainCategories[i] = domain.DishCategory{
			Id:   category.Id,
			Name: category.Name,
		}
	}
	return domainCategories, nil
}

func (s DishCategory) GetCategory(ctx context.Context, id int32) (domain.DishCategory, error) {
	category, err := s.repo.GetCategory(ctx, id)
	if err != nil {
		return domain.DishCategory{}, errors.WithMessage(err, "get category")
	}
	return domain.DishCategory{
		Id:   category.Id,
		Name: category.Name,
	}, nil
}

func (s DishCategory) AddCategory(ctx context.Context, category string) (int32, error) {
	id, err := s.repo.AddCategory(ctx, category)
	if err != nil {
		return 0, errors.WithMessage(err, "add category")
	}
	return id, nil
}

func (s DishCategory) RenameCategory(ctx context.Context, req domain.RenameCategoryRequest) error {
	err := s.repo.RenameCategory(ctx, req.Id, req.Name)
	if err != nil {
		return errors.WithMessage(err, "rename category")
	}
	return nil
}

func (s DishCategory) DeleteCategory(ctx context.Context, id int32) error {
	err := s.repo.DeleteCategory(ctx, id)
	if err != nil {
		return errors.WithMessage(err, "delete category")
	}
	return nil
}
