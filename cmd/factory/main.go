package main

import (
	"context"

	"github.com/aarondl/opt/null"
	"github.com/alexedwards/argon2id"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/db/models/factory"
	"github.com/tkahng/authgo/internal/tools/security"
	"github.com/tkahng/authgo/internal/tools/utils"
)

/**
f := factory.New()

f.AddBaseJetMods(
    factory.JetMods.RandomID(),
    factory.JetMods.RandomAirportID(),
)

// The jet templates will generate models with random IDs and AirportIDs
jetTemplate1 := f.NewJet()
jetTemplate2 := f.NewJet()

// We can also clear the base mods
f.ClearBaseJetMods()
// Create a new jet from the template
jet, err := jetTemplate.Create(ctx, db)
*/

func main() {
	ctx := context.Background()
	conf := conf.AppConfigGetter()

	dbx := core.NewBobFromConf(ctx, conf.Db)
	hash, err := security.CreateHash("password", argon2id.DefaultParams)
	if err != nil {
		panic(err)
	}
	f := factory.New()
	f.AddBaseUserMod(
		factory.UserMods.RandomEmail(nil),
		factory.UserMods.AddNewRoles(1,
			factory.RoleMods.AddNewPermissions(1,
				factory.PermissionMods.RandomName(nil),
			),
		),
		factory.UserMods.AddNewUserAccounts(1,
			factory.UserAccountMods.Provider(models.ProvidersCredentials),
			factory.UserAccountMods.Password(null.From(hash)),
		),
	)
	usertemplate := f.NewUser()
	jet, err := usertemplate.Create(ctx, dbx)
	if err != nil {
		panic(err)
	}
	accounts, err := jet.UserAccounts().All(ctx, dbx)
	if err != nil {
		panic(err)
	}
	roles, err := jet.Roles().All(ctx, dbx)
	if err != nil {
		panic(err)
	}
	// jet.LoadUserUserAccounts(ctx, dbx)
	utils.PrettyPrintJSON(jet)
	utils.PrettyPrintJSON(accounts)
	utils.PrettyPrintJSON(roles)
	// fmt.Println(utils.MarshalJSON(claims))
	// settings, err := models.AppParams.Insert(&models.AppParamSetter{
	// 	Name:  omit.From("settings"),
	// 	Value: omit.From(),
	// })
}
