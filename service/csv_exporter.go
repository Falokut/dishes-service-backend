package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"dishes-service-backend/entity"

	"github.com/pkg/errors"
)

type OrderExporterRepo interface {
	GetOrdersByPeriod(ctx context.Context, start time.Time, end time.Time) ([]entity.OrderToExport, error)
}

type CsvExporter struct {
	repo OrderExporterRepo
}

func NewCsvOrderExporter(repo OrderExporterRepo) CsvExporter {
	return CsvExporter{
		repo: repo,
	}
}

func (s CsvExporter) GetOrdersCsv(ctx context.Context, start time.Time, end time.Time) ([]byte, error) {
	orders, err := s.repo.GetOrdersByPeriod(ctx, start, end)
	if err != nil {
		return nil, errors.WithMessage(err, "get orders by period")
	}
	toExport := make([][]string, 0, len(orders)+1)
	toExport = append(toExport, []string{
		"\uFEFFномер заказа",
		"дата заказа",
		"статус",
		"метод оплаты",
		"telegram ник сотрудника",
		"стоимость заказа",
		"состав заказа",
	})
	for _, order := range orders {
		orderItems := make([]string, 0, len(order.Items))
		for _, item := range order.Items {
			orderItems = append(orderItems,
				fmt.Sprintf("блюдо: '%s' количество: %d итоговая цена: %s ресторан: %s",
					item.Name, item.Count, formatMoney(item.Price), item.RestaurantName,
				),
			)
		}

		toExport = append(toExport, []string{
			order.Id,
			order.CreatedAt.Format(entity.DataFormat),
			order.Status,
			order.PaymentMethod,
			order.Username,
			formatMoney(order.Total),
			strings.Join(orderItems, ","),
		})
	}
	var b bytes.Buffer
	writer := csv.NewWriter(&b)
	err = writer.WriteAll(toExport)
	if err != nil {
		return nil, errors.WithMessage(err, "write all")
	}
	writer.Flush()
	return b.Bytes(), nil
}

func formatMoney(money int32) string {
	return fmt.Sprintf("%d.%d₽", money/100, money%100) // nolint:mnd
}
