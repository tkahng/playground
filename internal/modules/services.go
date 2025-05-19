package modules

import (
	"github.com/tkahng/authgo/internal/modules/authmodule"
	"github.com/tkahng/authgo/internal/modules/paymentmodule"
	"github.com/tkahng/authgo/internal/modules/securitymodule"
	"github.com/tkahng/authgo/internal/modules/teammodule"
)

type Services struct {
	Jwt      securitymodule.JwtService
	Password securitymodule.PasswordService
	Payment  paymentmodule.PaymentService
	Team     teammodule.TeamService
	Auth     authmodule.AuthService
}

// func NewServices() *Services {
// 	jwt := securitymodule.NewJwtService()
// 	password := securitymodule.NewPasswordService()
// 	payment := paymentmodule.NewPaymentService(

// 	)
// }
