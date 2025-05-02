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

type Restaurant struct {
	cli db.DB
}

func NewRestaurant(cli db.DB) Restaurant {
	return Restaurant{cli: cli}
}
func (r Restaurant) GetAllRestaurants(ctx context.Context) ([]entity.Restaurant, error) {
	const query = "SELECT id, name FROM restaurants ORDER BY id"
	var restaurants []entity.Restaurant
	err := r.cli.Select(ctx, &restaurants, query)
	if err != nil {
		return nil, errors.WithMessagef(err, "exec query '%s'", query)
	}
	return restaurants, nil
}

func (r Restaurant) GetRestaurant(ctx context.Context, id int32) (entity.Restaurant, error) {
	const query = "SELECT id, name FROM restaurants WHERE id=$1"
	var restaurantName entity.Restaurant
	err := r.cli.SelectRow(ctx, &restaurantName, query, id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return entity.Restaurant{}, domain.ErrRestaurantNotFound
	case err != nil:
		return entity.Restaurant{}, errors.WithMessagef(err, "exec query '%s'", query)
	default:
		return restaurantName, nil
	}
}

func (r Restaurant) InsertRestaurant(ctx context.Context, restaurantName string) (int32, error) {
	query := `WITH e AS(
    INSERT INTO restaurants (name) 
           VALUES ($1) 
    ON CONFLICT DO NOTHING
    RETURNING id
	)
	SELECT * FROM e UNION SELECT id FROM restaurants WHERE name=$1;`

	var id int32
	err := r.cli.SelectRow(ctx, &id, query, restaurantName)
	if err != nil {
		return 0, errors.WithMessagef(err, "exec query '%s'", query)
	}
	return id, nil
}

func (r Restaurant) RenameRestaurant(ctx context.Context, id int32, newName string) error {
	const query = "UPDATE restaurants SET name = $1 WHERE id = $2"
	_, err := r.cli.Exec(ctx, query, newName, id)
	var pgErr *pgconn.PgError
	switch {
	case errors.As(err, &pgErr) && pgErr.SQLState() == pgerrcode.UniqueViolation:
		return domain.ErrRestaurantConflict
	case err != nil:
		return errors.WithMessagef(err, "exec query '%s'", query)
	default:
		return nil
	}
}

func (r Restaurant) DeleteRestaurant(ctx context.Context, id int32) error {
	const query = "DELETE FROM restaurants WHERE id=$1"
	_, err := r.cli.Exec(ctx, query, id)
	if err != nil {
		return errors.WithMessagef(err, "exec query '%s'", query)
	}
	return nil
}
