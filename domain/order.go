package domain

import "time"

type ProcessOrderRequest struct {
	Items         map[string]int32 `validate:"required"`
	PaymentMethod string           `validate:"required,min=1"`
	Wishes        string           `json:",omitempty"`
}

type ProcessOrderResponse struct {
	// for some payment methods may be empty
	PaymentUrl string
}

type GetMyOrdersRequest struct {
	Limit  int32 `validate:"min=0,max=40"`
	Offset int32 `validate:"min=0"`
}

type UserOrder struct {
	Id            string
	Items         []OrderItem
	PaymentMethod string
	Total         int32
	Status        string
	Wishes        string `json:",omitempty"`
	CreatedAt     time.Time
}

type OrderItem struct {
	DishId     int32
	Name       string
	Price      int32
	Count      int32
	TotalPrice int32
	Status     string
}
