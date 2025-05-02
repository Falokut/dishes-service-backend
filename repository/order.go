package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"dishes-service-backend/domain"
	"dishes-service-backend/entity"

	"github.com/Falokut/go-kit/db"
	"github.com/pkg/errors"
)

type Order struct {
	cli db.DB
}

func NewOrder(cli db.DB) Order {
	return Order{
		cli: cli,
	}
}

func (r Order) InsertOrder(ctx context.Context, order *entity.Order) error {
	query := `INSERT INTO 
	orders(id, user_id, total, created_at, wishes, payment_method, status)
	VALUES($1,$2,$3,$4,$5,$6,$7)`
	_, err := r.cli.Exec(ctx, query,
		order.Id,
		order.UserId,
		order.Total,
		order.CreatedAt,
		order.Wishes,
		order.PaymentMethod,
		order.Status,
	)
	if err != nil {
		return errors.WithMessagef(err, "exec query '%s'", query)
	}
	return nil
}

//nolint:mnd
func (r Order) InsertOrderItems(ctx context.Context, orderId string, items entity.OrderItems) error {
	args := make([]any, 0, len(items)*3+1)
	args = append(args, orderId)
	placeholders := make([]string, len(items))
	for i, item := range items {
		placeholders[i] = fmt.Sprintf("($1,$%d,$%d,$%d)",
			len(args)+1,
			len(args)+2,
			len(args)+3,
		)
		args = append(args, item.DishId, item.Count, item.Price)
	}

	query := fmt.Sprintf(`INSERT INTO order_items(order_id,dish_id,count,price) VALUES %s`,
		strings.Join(placeholders, ","))
	_, err := r.cli.Exec(ctx, query, args...)
	if err != nil {
		return errors.WithMessagef(err, "exec query '%s'", query)
	}
	return nil
}

func (r Order) UpdateOrderStatus(ctx context.Context, orderId, newStatus string) error {
	query := "UPDATE orders SET status=$1 WHERE id=$2"
	_, err := r.cli.Exec(ctx, query, newStatus, orderId)
	if err != nil {
		return errors.WithMessagef(err, "exec query '%s'", query)
	}
	return nil
}

func (r Order) GetOrder(ctx context.Context, orderId string) (*entity.Order, error) {
	query := `
	SELECT
		o.id,
		o.payment_method,
		o.user_id,
		o.total, 
		o.created_at,
		o.wishes,
		o.status,
		json_agg(
			json_build_object(
			'dishId', oi.dish_id,
			'count', oi.count,
			'price', oi.price,
			'restaurantName', r.name,
			'name', d.name
			)
		) AS items
    FROM orders o
    JOIN order_items oi ON o.id = oi.order_id
	JOIN dish d ON oi.dish_id = d.id
	JOIN restaurants AS r ON d.restaurant_id = r.id
    WHERE o.id = $1
	GROUP BY o.id`

	var order entity.Order
	err := r.cli.SelectRow(ctx, &order, query, orderId)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, domain.ErrOrderNotFound
	case err != nil:
		return nil, errors.WithMessagef(err, "exec query '%s'", query)
	default:
		return &order, nil
	}
}

func (r Order) GetOrderStatus(ctx context.Context, orderId string) (string, error) {
	query := "SELECT status FROM orders WHERE id=$1"
	var status string
	err := r.cli.SelectRow(ctx, &status, query, orderId)
	if err != nil {
		return "", errors.WithMessagef(err, "exec query '%s'", query)
	}
	return status, nil
}

func (r Order) SetOrderStatus(ctx context.Context, orderId, oldStatus, newStatus string) error {
	query := "UPDATE orders SET status=$1 WHERE id=$2 AND status=$3"
	_, err := r.cli.Exec(ctx, query, newStatus, orderId, oldStatus)
	if err != nil {
		return errors.WithMessagef(err, "exec query '%s'", query)
	}
	return nil
}

func (r Order) InsertAllowOrderingAudit(ctx context.Context) error {
	const query = "INSERT INTO allow_ordering_audit DEFAULT VALUES"
	_, err := r.cli.Exec(ctx, query)
	if err != nil {
		return errors.WithMessagef(err, "exec query '%s'", query)
	}
	return nil
}

func (r Order) SetOrderingAuditEndPeriod(ctx context.Context) error {
	const query = "UPDATE allow_ordering_audit SET end_period=$1 WHERE end_period IS NULL"
	_, err := r.cli.Exec(ctx, query, time.Now())
	if err != nil {
		return errors.WithMessagef(err, "exec query '%s'", query)
	}
	return nil
}

func (r Order) IsOrderingAllowed(ctx context.Context) (bool, error) {
	const query = "SELECT EXISTS(SELECT id FROM allow_ordering_audit WHERE end_period IS NULL)"
	isAllowed := false
	err := r.cli.SelectRow(ctx, &isAllowed, query)
	if err != nil {
		return false, errors.WithMessagef(err, "exec query '%s'", query)
	}
	return isAllowed, nil
}

func (r Order) GetUserOrders(ctx context.Context, userId string, limit int32, offset int32) ([]entity.Order, error) {
	query := `
	SELECT
		o.id,
		o.payment_method,
		o.user_id,
		o.total, 
		o.created_at,
		o.wishes,
		o.status,
		json_agg(
			json_build_object(
			'dishId', oi.dish_id,
			'count', oi.count,
			'price', oi.price,
			'name', d.name
			)
		) AS items
    FROM orders o
    JOIN order_items oi ON o.id = oi.order_id
	JOIN dish d ON oi.dish_id = d.id
    WHERE o.user_id = $1
	GROUP BY o.id
    ORDER BY o.created_at DESC
	LIMIT $2
	OFFSET $3`
	var orders []entity.Order
	err := r.cli.Select(ctx, &orders, query, userId, limit, offset)
	if err != nil {
		return nil, errors.WithMessagef(err, "exec query '%s'", query)
	}
	return orders, nil
}

func (r Order) GetOrderedChatId(ctx context.Context, orderId string) (int64, error) {
	query := `
	SELECT chat_id 
	FROM orders o
	JOIN users_telegrams ut ON o.user_id=ut.id
	WHERE o.id=$1`
	var chatId int64
	err := r.cli.SelectRow(ctx, &chatId, query, orderId)
	if err != nil {
		return 0, errors.WithMessagef(err, "exec query '%s'", query)
	}
	return chatId, nil
}

func (r Order) GetOrdersByPeriod(ctx context.Context, start time.Time, end time.Time) ([]entity.OrderToExport, error) {
	query := `
	SELECT
		o.id,
		o.payment_method,
		u.username,
		o.total, 
		o.created_at,
		o.status,
		json_agg(
			json_build_object(
			'dishId', oi.dish_id,
			'count', oi.count,
			'price', oi.price,
			'name', d.name
			)
		) AS items
    FROM orders o
    JOIN order_items oi ON o.id = oi.order_id
	JOIN dish d ON oi.dish_id = d.id
	JOIN users u ON o.user_id = u.id
    WHERE o.created_at >= $1 AND o.created_at <= $2
	GROUP BY o.id, u.username
    ORDER BY o.created_at DESC`
	var orders []entity.OrderToExport
	err := r.cli.Select(ctx, &orders, query, start, end)
	if err != nil {
		return nil, errors.WithMessagef(err, "exec query '%s'", query)
	}
	return orders, nil
}
