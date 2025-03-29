package conf

import (
	"github.com/caarlos0/env/v11"
)

type ConfigGetter func() EnvConfig

func AppConfigGetter() EnvConfig {
	var config EnvConfig
	if err := env.ParseWithOptions(&config, env.Options{
		RequiredIfNoDef: true,
	}); err != nil {
		panic(err)
	}
	// db = &config.Db
	// auth = &config.Auth
	// mail = &config.Mail
	// oauth = &config.OAuth
	// stripe = &config.Stripe
	// common = &config.Common
	// ai = &config.Ai
	// if db == nil || auth == nil || mail == nil || oauth == nil || stripe == nil || common == nil || ai == nil {
	// 	log.Println(config)
	// 	panic("Unable to load config")
	// }
	return config
}
