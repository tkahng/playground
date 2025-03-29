package migrator

import (
	"fmt"
	"net/url"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	database "github.com/tkahng/authgo/db"
)

func Migrate(uri string) {
	u, _ := url.Parse(uri)
	db := dbmate.New(u)
	db.FS = database.Migrations
	db.MigrationsDir = []string{"./migrations"}
	db.Drop()
	db.Create()
	fmt.Println("Migrations:")
	migrations, err := db.FindMigrations()
	if err != nil {
		panic(err)
	}
	for _, m := range migrations {
		fmt.Println(m.Version, m.FilePath)
	}
	fmt.Println("\nApplying...")
	err = db.Migrate()
	if err != nil {
		panic(err)
	}
}
