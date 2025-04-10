package main

import (
	"context"
	"fmt"

	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/tools/utils"
)

type AuthenticationClaims struct {
	// jwt.RegisteredClaims
	// Type   shared.TokenType `json:"type"`
	UserId string `json:"user_id"`
	Email  string `json:"email"`
}

func main() {

	// fmt.Print(query)
	ctx := context.Background()
	conf := conf.AppConfigGetter()

	dbx := core.NewPoolFromConf(ctx, conf.Db)
	db := core.NewBobFromPool(dbx)
	res, err := repository.GetUserWithRolesAndPermissions(ctx, db, "tkahng@gmail.com")
	if err != nil {
		fmt.Println(err)
		return
	}
	utils.PrettyPrintJSON(res)
	// shouldReturn := newFunction(ctx, db)
	// if shouldReturn {
	// 	return
	// }
	// fmt.Println(utils.MarshalJSON(claims))
	// settings, err := models.AppParams.Insert(&models.AppParamSetter{
	// 	Name:  omit.From("settings"),
	// 	Value: omit.From(),
	// })
}
