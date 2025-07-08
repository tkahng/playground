package apis

import (
	"github.com/tkahng/authgo/internal/core"
)

type Api struct {
	app core.App
}

func (a *Api) App() core.App {
	if a.app == nil {
		panic("app not initialized for api")
	}
	return a.app
}

func NewApi(app core.App) *Api {
	return &Api{
		app: app,
	}
}

type ApiDecorator struct {
	*Api
}

func NewApiDecorator(app core.App) *ApiDecorator {
	return &ApiDecorator{
		Api: NewApi(app),
	}
}
