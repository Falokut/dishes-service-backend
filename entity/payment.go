package entity

type ProcessPaymentRequest struct {
	UserId  string
	OrderId string
	Total   int32
}
