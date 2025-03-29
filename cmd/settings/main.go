package main

import (
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/tools/utils"
)

func main() {

	// fmt.Print(query)
	// ctx := context.Background()
	// conf := conf.AppConfigGetter()

	settings := core.NewDefaultSettings()
	utils.PrettyPrintJSON(settings)

}
