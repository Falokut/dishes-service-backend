package payment

import (
	"dishes-service-backend/repository"

	telegram_payment "dishes-service-backend/service/payment/telegram"
	"github.com/txix-open/bgjob"
)

func NewPaymentMethods(userRepo repository.User, bgJobCli *bgjob.Client) map[string]PaymentService {
	return map[string]PaymentService{
		telegram_payment.PaymentMethod: telegram_payment.NewPayment(userRepo, bgJobCli),
	}
}
