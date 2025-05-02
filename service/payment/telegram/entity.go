package payment

import (
	"dishes-service-backend/entity"
)

type PaymentPayload struct {
	Order  entity.Order
	ChatId int64
}
