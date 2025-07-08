package apis

import (
	"github.com/tkahng/authgo/internal/core"
)

type Api struct {
	a core.App
}

func (a *Api) App() core.App {
	if a.a == nil {
		panic("app not initialized for api")
	}
	return a.a
}

func NewApi(app core.App) *Api {
	return &Api{
		a: app,
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
