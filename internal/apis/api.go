package apis

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/httprate"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/tools/logger"
	"github.com/tkahng/playground/internal/tools/types"
	"github.com/tkahng/playground/ui"
)

type AppApi interface {
	Api() huma.API
	RegisterRoutes()
	Router() chi.Router
	App() core.App
}

type Api struct {
	app    core.App
	api    huma.API
	router chi.Router
}

var _ AppApi = (*Api)(nil)

func (appApi *Api) RegisterRoutes() {
	bindMiddlewares(appApi.Api(), appApi.App())
	bindApis(appApi.Api(), appApi)
}

func (a *Api) Api() huma.API {
	if a.api == nil {
		panic("api not initialized for api")
	}
	return a.api
}

func (a *Api) Router() chi.Router {
	if a.router == nil {
		panic("router not initialized for api")
	}
	return a.router
}

func (a *Api) App() core.App {
	if a.app == nil {
		panic("app not initialized for api")
	}
	return a.app
}

func NewAppApiWithRouter(app core.App) *Api {
	router := NewRouter(app)
	api := NewApiGroup(router)
	return &Api{
		app:    app,
		api:    api,
		router: router,
	}
}
func NewAppApi(app core.App, router chi.Router, api huma.API) *Api {
	return &Api{
		app:    app,
		api:    api,
		router: router,
	}
}

func NewRouter(app core.App) *chi.Mux {
	r := chi.NewMux()
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
	}))
	r.Use(httplog.RequestLogger(app.Logger(), &httplog.Options{
		// Level defines the verbosity of the request logs:
		// slog.LevelDebug - log all responses (incl. OPTIONS)
		// slog.LevelInfo  - log responses (excl. OPTIONS)
		// slog.LevelWarn  - log 4xx and 5xx responses only (except for 429)
		// slog.LevelError - log 5xx responses only
		Level: slog.LevelInfo,

		// Set log output to Elastic Common Schema (ECS) format.
		Schema: logger.GetDefaultFormat(&app.Config().AppConfig),

		// RecoverPanics recovers from panics occurring in the underlying HTTP handlers
		// and middlewares. It returns HTTP 500 unless response status was already set.
		//
		// NOTE: Panics are logged as errors automatically, regardless of this setting.
		RecoverPanics: true,
	}))
	// r.Use(middleware.Logger)
	// r.Use(middleware.Recoverer)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))
	// Handle all other routes by serving index.html (for React Router)
	r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		p := filepath.Clean(r.URL.Path)
		if strings.Contains(p, ".") {
			http.FileServer(http.FS(ui.DistDirFS)).ServeHTTP(w, r)
			return
		}
		if _, err := ui.DistDirFS.Open(p); err != nil {
			file, err := ui.DistDirFS.Open("index.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer file.Close()

			ff, ok := file.(io.ReadSeeker)
			if !ok {
				http.Error(w, "[FileFS] file does not implement io.ReadSeeker", http.StatusInternalServerError)
				return
			}

			http.ServeContent(w, r, "index.html", time.Now(), ff)
		} else {
			http.FileServer(http.FS(ui.DistDirFS)).ServeHTTP(w, r)
		}
	})

	return r
}

func NewApiGroup(r chi.Router) huma.API {
	var api huma.API
	config := huma.DefaultConfig("My API", "1.0.0")
	config.Servers = []*huma.Server{{URL: "http://localhost:8080"}}
	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		shared.BearerAuthSecurityKey: {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
	}

	api = humachi.New(r, config)

	grp := huma.NewGroup(api, "/api")
	return grp
}

func bindMiddlewares(api huma.API, app core.App) {
	api.UseMiddleware(humamiddleware.HumaChiMiddleware(middleware.RecovererMiddleware(app)))
	api.UseMiddleware(humamiddleware.HumaAuthMiddleware(api, app))
	api.UseMiddleware(humamiddleware.HumaRequireAuthMiddleware(api, app))
}

type IndexOutputBody struct {
	Access string `json:"access"`
}

type IndexOutput struct {
	Body IndexOutputBody `json:"body"`
}

func bindApis(api huma.API, appApi *Api) {

	// Misc routes ------------------------------------
	BindMiscApi(api, appApi)
	// signup -------------------------------------------------------------
	BindAuthApi(api, appApi)

	// ---- Upload File
	BindMediaApi(api, appApi)

	// ---- Teams
	BindTeamsApi(api, appApi)

	// stats routes -------------------------------------------------------------------------------------------------
	BindStatsApi(api, appApi)

	// ---- task routes -------------------------------------------------------------------------------------------------
	BindTaskApi(api, appApi)

	// stripe routes -------------------------------------------------------------------------------------------------

	BindStripeApi(api, appApi)

	//  admin routes ----------------------------------------------------------------------------
	BindAdminApi(api, appApi)
	// admin stripe products with prices
	BindUserReactionApi(api, appApi)
}

func BindMiscApi(api huma.API, appApi *Api) {
	huma.Get(api, "/", func(ctx context.Context, input *struct {
		Page types.OmittableNullable[string] `query:"page" required:"false"`
	}) (*IndexOutput, error) {
		fmt.Println("input", input)
		return &IndexOutput{
			Body: IndexOutputBody{
				Access: "public",
			},
		}, nil
	})

	//  public list of permissions -----------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "permissions-list",
			Method:      http.MethodGet,
			Path:        "/permissions",
			Summary:     "permissions list",
			Description: "List of permissions",
			Tags:        []string{"Permissions"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.PermissionsList,
	)
	// protected test routes -----------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "api-protected",
			Method:      http.MethodGet,
			Path:        "/protected/{permission-name}",
			Summary:     "Api protected",
			Description: "Api protected",
			Tags:        []string{"Protected"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.ApiProtected,
	)
}
