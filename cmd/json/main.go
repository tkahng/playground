package main

import (
	"github.com/tkahng/authgo/internal/tools/utils"
)

type AuthenticationClaims struct {
	// jwt.RegisteredClaims
	// Type   shared.TokenType `json:"type"`
	UserId string `json:"user_id"`
	Email  string `json:"email"`
}

func main() {
	claims := AuthenticationClaims{
		UserId: "1",
		Email:  "email",
	}
	utils.PrettyPrintJSON(claims)

	claims.Email = "email2"
	utils.PrettyPrintJSON(claims)
	// fmt.Println(utils.MarshalJSON(claims))
	// settings, err := models.AppParams.Insert(&models.AppParamSetter{
	// 	Name:  omit.From("settings"),
	// 	Value: omit.From(),
	// })
}
