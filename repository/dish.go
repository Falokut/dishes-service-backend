package repository

import (
	"context"
	"fmt"
	"strings"

	"dishes-service-backend/entity"

	"github.com/Falokut/go-kit/db"
	"github.com/pkg/errors"
)

type Dish struct {
	cli db.DB
}

func NewDish(cli db.DB) Dish {
	return Dish{
		cli: cli,
	}
}

func (r Dish) List(ctx context.Context, limit, offset int32) ([]entity.Dish, error) {
	query := `
	SELECT
		d.id,
		d.name,
		d.description,
	    d.price, 
		COALESCE(d.image_id,'') AS image_id,
		array_to_string(ARRAY_AGG(COALESCE(c.name,'')),',') AS categories,
		r.name AS restaurant_name 
	FROM dish AS d
	JOIN restaurants AS r ON d.restaurant_id = r.id
	LEFT JOIN dish_categories AS f_c ON d.id=f_c.dish_id
	LEFT JOIN categories AS c ON f_c.category_id=c.id
	GROUP BY d.id, d.name, d.description, d.price, d.image_id, r.name
	ORDER BY d.id
	LIMIT $1 OFFSET $2`
	var res []entity.Dish
	err := r.cli.Select(ctx, &res, query, limit, offset)
	if err != nil {
		return nil, errors.WithMessage(err, "get dish list")
	}
	return res, nil
}

func (r Dish) InsertDish(ctx context.Context, req *entity.InsertDish) (int32, error) {
	query := `INSERT INTO dish
	(name, description, price, image_id, restaurant_id) 
	VALUES($1, $2, $3, $4, $5)
	RETURNING id;`
	var id int32
	err := r.cli.SelectRow(ctx, &id, query, req.Name, req.Description, req.Price, req.ImageId, req.RestaurantId)
	if err != nil {
		return 0, errors.WithMessagef(err, "exec query '%s'", query)
	}
	return id, nil
}

func (r Dish) InsertDishCategories(ctx context.Context, dishId int32, categories []int32) error {
	query, args := getInsertDishCategoriesQuery(dishId, categories)
	if len(args) == 0 {
		return nil
	}
	_, err := r.cli.Exec(ctx, query, args...)
	if err != nil {
		return errors.WithMessagef(err, "exec query '%s'", query)
	}
	return nil
}

func (r Dish) GetDishesByIds(ctx context.Context, ids []int32) ([]entity.Dish, error) {
	query := `
	SELECT 
		d.id,
		d.name,
		d.description,
		d.price,
		COALESCE(d.image_id,'') AS image_id,
		array_to_string(ARRAY_AGG(COALESCE(c.name,'')),',') AS categories,
		r.name AS restaurant_name 
	FROM dish AS d
	JOIN restaurants AS r ON d.restaurant_id = r.id
	LEFT JOIN dish_categories AS f_c ON d.id=f_c.dish_id
	LEFT JOIN categories AS c ON f_c.category_id=c.id
	WHERE d.id=ANY($1)
	GROUP BY d.id, d.name, d.description, d.price, d.image_id, r.name
	ORDER BY d.id;`

	var res []entity.Dish
	err := r.cli.Select(ctx, &res, query, ids)
	if err != nil {
		return nil, errors.WithMessage(err, "get dish list")
	}
	return res, nil
}

func (r Dish) GetDishesByCategories(ctx context.Context, limit int32, offset int32, ids []int32) ([]entity.Dish, error) {
	query := `
	SELECT 
		d.id,
		d.name,
		d.description,
		d.price,
		COALESCE(d.image_id,'') AS image_id,
		array_to_string(ARRAY_AGG(COALESCE(c.name,'')),',') AS categories,
		r.name AS restaurant_name 
	FROM dish AS d
	JOIN restaurants AS r ON d.restaurant_id = r.id
	LEFT JOIN dish_categories AS f_c ON d.id=f_c.dish_id
	LEFT JOIN categories AS c ON f_c.category_id=c.id
	GROUP BY d.id, d.name, d.description, d.price, d.image_id, r.name
	HAVING array_agg(c.id) @> $1
	ORDER BY d.id
	LIMIT $2 OFFSET $3;`

	var res []entity.Dish
	err := r.cli.Select(ctx, &res, query, ids, limit, offset)
	if err != nil {
		return nil, errors.WithMessage(err, "get dish list")
	}
	return res, nil
}

func (r Dish) EditDish(ctx context.Context, req *entity.EditDish) error {
	query := `UPDATE dish SET name=$1, description=$2, price=$3, image_id=$4, restaurant_id=$5 WHERE id=$6`
	_, err := r.cli.Exec(ctx, query, req.Name, req.Description, req.Price, req.ImageId, req.RestaurantId, req.Id)
	if err != nil {
		return errors.WithMessagef(err, "exec query '%s'", query)
	}
	return nil
}

func (r Dish) DeleteDishCategories(ctx context.Context, dishId int32) error {
	const query = "DELETE FROM dish_categories WHERE dish_id=$1;"
	_, err := r.cli.Exec(ctx, query, dishId)
	if err != nil {
		return errors.WithMessagef(err, "exec query '%s'", query)
	}
	return nil
}

func (r Dish) DeleteDish(ctx context.Context, id int32) error {
	_, err := r.cli.Exec(ctx, "DELETE FROM dish WHERE id=$1", id)
	if err != nil {
		return errors.WithMessage(err, "delete dishes")
	}
	return nil
}

func getInsertDishCategoriesQuery(id int32, categories []int32) (string, []any) {
	if len(categories) == 0 {
		return "", nil
	}
	var valuesPlaceholders = make([]string, len(categories))
	var args = make([]any, 0, len(categories)+1)
	args = append(args, id)
	for i, catId := range categories {
		valuesPlaceholders[i] = fmt.Sprintf("($1,$%d)", len(args)+1)
		args = append(args, catId)
	}
	return fmt.Sprintf(`INSERT INTO dish_categories(dish_id, category_id) VALUES %s ON CONFLICT DO NOTHING`,
		strings.Join(valuesPlaceholders, ",")), args
}
