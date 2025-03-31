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
	return config
}
