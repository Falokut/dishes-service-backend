package service

import (
	"context"
	"strings"

	"github.com/Falokut/go-kit/log"
	"github.com/google/uuid"

	"dishes-service-backend/domain"
	"dishes-service-backend/entity"

	"github.com/pkg/errors"
)

type DishRepo interface {
	List(ctx context.Context, limit, offset int32) ([]entity.Dish, error)
	GetDishesByIds(ctx context.Context, ids []int32) ([]entity.Dish, error)
	GetDishesByCategories(ctx context.Context, limit int32, offset int32, ids []int32) ([]entity.Dish, error)
	DeleteDish(ctx context.Context, id int32) error
}

type FileRepo interface {
	UploadFile(ctx context.Context, req entity.UploadFileRequest) error
	DeleteFile(ctx context.Context, category string, imageId string) error
	GetFileUrl(category, imageId string) string
}

type AddDishTx interface {
	InsertDish(ctx context.Context, req *entity.InsertDish) (int32, error)
	InsertDishCategories(ctx context.Context, dishId int32, categories []int32) error
}

type EditDishTx interface {
	InsertDishCategories(ctx context.Context, dishId int32, categories []int32) error
	EditDish(ctx context.Context, req *entity.EditDish) error
	DeleteDishCategories(ctx context.Context, dishId int32) error
}

type DishTxRunner interface {
	AddDishTx(ctx context.Context, tx func(ctx context.Context, tx AddDishTx) error) error
	EditDishTx(ctx context.Context, tx func(ctx context.Context, tx EditDishTx) error) error
}

const dishImageCategory = "image-dish"

type Dish struct {
	dishRepo DishRepo
	txRunner DishTxRunner
	fileRepo FileRepo
	logger   log.Logger
}

func NewDish(
	dishRepo DishRepo,
	txRunner DishTxRunner,
	fileRepo FileRepo,
	logger log.Logger,
) Dish {
	return Dish{
		dishRepo: dishRepo,
		txRunner: txRunner,
		fileRepo: fileRepo,
		logger:   logger,
	}
}

func (s Dish) List(ctx context.Context, limit, offset int32) ([]domain.Dish, error) {
	if limit == 0 {
		limit = 30
	}
	dish, err := s.dishRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, errors.WithMessage(err, "dish list")
	}
	converted := make([]domain.Dish, len(dish))
	for i, f := range dish {
		converted[i] = s.dishFromEntity(f)
	}
	return converted, nil
}

func (s Dish) GetByIds(ctx context.Context, ids []int32) ([]domain.Dish, error) {
	dish, err := s.dishRepo.GetDishesByIds(ctx, ids)
	if err != nil {
		return nil, errors.WithMessage(err, "dish list by ids")
	}
	converted := make([]domain.Dish, len(dish))
	for i, f := range dish {
		converted[i] = s.dishFromEntity(f)
	}
	return converted, nil
}

func (s Dish) GetByCategories(ctx context.Context, limit, offset int32, ids []int32) ([]domain.Dish, error) {
	dish, err := s.dishRepo.GetDishesByCategories(ctx, limit, offset, ids)
	if err != nil {
		return nil, errors.WithMessage(err, "dish list by ids")
	}
	converted := make([]domain.Dish, len(dish))
	for i, f := range dish {
		converted[i] = s.dishFromEntity(f)
	}
	return converted, nil
}

func (s Dish) AddDish(ctx context.Context, req domain.AddDishRequest) (*domain.AddDishResponse, error) {
	var dishId int32
	var err error
	err = s.txRunner.AddDishTx(ctx, func(ctx context.Context, tx AddDishTx) error {
		dishId, err = s.addDish(ctx, req, tx)
		if err != nil {
			return errors.WithMessage(err, "add dish")
		}
		return nil
	})
	if err != nil {
		return nil, errors.WithMessage(err, "add dish tx")
	}
	return &domain.AddDishResponse{Id: dishId}, nil
}

func (s Dish) addDish(ctx context.Context, req domain.AddDishRequest, tx AddDishTx) (int32, error) {
	imageId := ""
	if len(req.Image) > 0 {
		imageId = uuid.NewString()
	}
	dishId, err := tx.InsertDish(ctx, &entity.InsertDish{
		Name:         req.Name,
		Description:  req.Description,
		Price:        req.Price,
		ImageId:      imageId,
		RestaurantId: req.RestaurantId,
	})
	if err != nil {
		return 0, errors.WithMessage(err, "insert dish")
	}

	if len(req.Categories) > 0 {
		err = tx.InsertDishCategories(ctx, dishId, req.Categories)
		if err != nil {
			return 0, errors.WithMessage(err, "insert dish categories")
		}
	}
	if len(req.Image) > 0 {
		err = s.fileRepo.UploadFile(ctx, entity.UploadFileRequest{
			Category: dishImageCategory,
			Filename: imageId,
			Content:  req.Image,
		})
		if err != nil {
			return 0, errors.WithMessage(err, "upload file")
		}
	}

	return dishId, nil
}

func (s Dish) EditDish(ctx context.Context, req domain.EditDishRequest) error {
	err := s.txRunner.EditDishTx(ctx, func(ctx context.Context, tx EditDishTx) error {
		err := s.editDish(ctx, req, tx)
		if err != nil {
			return errors.WithMessage(err, "edit dish")
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "edit dish tx")
	}
	return nil
}

func (s Dish) editDish(ctx context.Context, req domain.EditDishRequest, tx EditDishTx) error {
	imageId := ""
	if len(req.Image) > 0 {
		imageId = uuid.NewString()
	}

	err := tx.EditDish(ctx, &entity.EditDish{
		Id:           req.Id,
		Name:         req.Name,
		Description:  req.Description,
		Price:        req.Price,
		ImageId:      imageId,
		RestaurantId: req.RestaurantId,
	})
	if err != nil {
		return errors.WithMessage(err, "edit dish")
	}

	err = tx.DeleteDishCategories(ctx, req.Id)
	if err != nil {
		return errors.WithMessage(err, "delete dish categories")
	}

	if len(req.Categories) > 0 {
		err = tx.InsertDishCategories(ctx, req.Id, req.Categories)
		if err != nil {
			return errors.WithMessage(err, "insert dish categories")
		}
	}

	err = s.deleteDishImage(ctx, req.Id)
	if err != nil {
		return errors.WithMessage(err, "delete dish image")
	}

	if len(req.Image) > 0 {
		err = s.fileRepo.UploadFile(ctx, entity.UploadFileRequest{
			Category: dishImageCategory,
			Filename: imageId,
			Content:  req.Image,
		})
	}

	if err != nil {
		return errors.WithMessage(err, "upload file")
	}
	return nil
}

func (s Dish) DeleteDish(ctx context.Context, id int32) error {
	err := s.deleteDishImage(ctx, id)
	if err != nil {
		return errors.WithMessage(err, "delete dish image")
	}

	err = s.dishRepo.DeleteDish(ctx, id)
	if err != nil {
		return errors.WithMessage(err, "delete dish categories")
	}

	return nil
}

func (s Dish) deleteDishImage(ctx context.Context, dishId int32) error {
	dishes, err := s.dishRepo.GetDishesByIds(ctx, []int32{dishId})
	if err != nil {
		return errors.WithMessage(err, "get dishes by ids")
	}
	if len(dishes) == 0 {
		return domain.ErrDishNotFound
	}

	dish := dishes[0]
	if dish.ImageId == "" {
		return nil
	}

	err = s.fileRepo.DeleteFile(ctx, dishImageCategory, dish.ImageId)
	if err != nil {
		s.logger.Warn(ctx, "delete image",
			log.String("imageId", dish.ImageId),
			log.String("category", dishImageCategory),
			log.Error(err),
		)
		return errors.WithMessage(err, "delete dish image")
	}
	return nil
}

func (s Dish) dishFromEntity(dish entity.Dish) domain.Dish {
	categories := []string{}
	if dish.Categories != "" {
		categories = strings.Split(dish.Categories, ",")
	}
	return domain.Dish{
		Id:             dish.Id,
		Name:           dish.Name,
		Description:    dish.Description,
		Price:          dish.Price,
		Url:            s.fileRepo.GetFileUrl(dishImageCategory, dish.ImageId),
		Categories:     categories,
		RestaurantName: dish.RestaurantName,
	}
}
