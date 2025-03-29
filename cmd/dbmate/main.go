package main

import (
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/db/migrator"
)

func main() {
	// wd, err := os.Getwd()
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// }
	// fmt.Println("Current working directory:", wd)
	cfg := conf.AppConfigGetter()
	migrator.Migrate(cfg.Db.DatabaseUrl)
}
