package main

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/tools/security"
	"github.com/tkahng/authgo/internal/tools/utils"
)

// type JwtClaimsA struct {
// 	jwt.RegisteredClaims
// 	Type       shared.TokenType `json:"type"`
// 	Provider   models.Providers `json:"provider"`
// 	RedirectTo string           `json:"redirect_to,omitempty"`
// }

// type TokenClaimsService interface {
// 	ParseToken(tokenString string, verificationKey string) (*jwt.MapClaims, error)
// 	ParseClaims(token *jwt.Token) (*JwtClaimsA, error)
// }

func main() {

	settings := core.DefaultAuthSettings()

	token, err := core.CreateOtpToken(&core.OtpPayload{
		UserId:     uuid.New(),
		Email:      "email@example.com",
		Token:      security.GenerateTokenKey(),
		Otp:        security.GenerateOtp(6),
		RedirectTo: "",
	}, settings.VerificationToken)

	fmt.Println(token, err)

	mapCLaims, err := security.ParseJWTMapClaims(token, settings.VerificationToken.Secret)
	if err != nil {
		fmt.Println(err)
		return
	}

	utils.PrettyPrintJSON(mapCLaims)

}
