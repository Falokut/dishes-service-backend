package repository

import (
	"context"
	"database/sql"
	"dishes-service-backend/domain"
	"dishes-service-backend/entity"

	"github.com/Falokut/go-kit/db"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
)

type DishCategory struct {
	cli db.DB
}

func NewDishCategory(cli db.DB) DishCategory {
	return DishCategory{cli: cli}
}
func (r DishCategory) GetAllCategories(ctx context.Context) ([]entity.DishCategory, error) {
	var categories []entity.DishCategory
	err := r.cli.Select(ctx, &categories, "SELECT id, name FROM categories ORDER BY id")
	if err != nil {
		return nil, errors.WithMessage(err, "execute query")
	}
	return categories, nil
}

func (r DishCategory) GetDishCategory(ctx context.Context) ([]entity.DishCategory, error) {
	var categories []entity.DishCategory
	query := `
	SELECT DISTINCT c.id, c.name
	FROM dish_categories dc
	JOIN categories c ON dc.category_id = c.id
	ORDER BY c.id;`
	err := r.cli.Select(ctx, &categories, query)
	if err != nil {
		return nil, errors.WithMessage(err, "execute query")
	}
	return categories, nil
}

func (r DishCategory) GetCategory(ctx context.Context, id int32) (entity.DishCategory, error) {
	var category entity.DishCategory
	err := r.cli.SelectRow(ctx, &category, "SELECT id, name FROM categories WHERE id=$1", id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return entity.DishCategory{}, domain.ErrDishCategoryNotFound
	case err != nil:
		return entity.DishCategory{}, errors.WithMessage(err, "execute query")
	default:
		return category, nil
	}
}

func (r DishCategory) AddCategory(ctx context.Context, category string) (int32, error) {
	query := `WITH e AS(
    INSERT INTO categories (name) 
           VALUES ($1) 
    ON CONFLICT DO NOTHING
    RETURNING id
	)
	SELECT * FROM e UNION SELECT id FROM categories WHERE name=$1;`

	var id int32
	err := r.cli.SelectRow(ctx, &id, query, category)
	if err != nil {
		return 0, errors.WithMessage(err, "execute query")
	}
	return id, nil
}

func (r DishCategory) RenameCategory(ctx context.Context, id int32, newName string) error {
	_, err := r.cli.Exec(ctx, "UPDATE categories SET name = $1 WHERE id = $2", newName, id)
	var pgErr *pgconn.PgError
	switch {
	case errors.As(err, &pgErr) && pgErr.SQLState() == pgerrcode.UniqueViolation:
		return domain.ErrDishCategoryConflict
	case err != nil:
		return errors.WithMessage(err, "execute query")
	default:
		return nil
	}
}

func (r DishCategory) DeleteCategory(ctx context.Context, id int32) error {
	_, err := r.cli.Exec(ctx, "DELETE FROM categories WHERE id=$1", id)
	if err != nil {
		return errors.WithMessage(err, "execute query")
	}
	return nil
}
