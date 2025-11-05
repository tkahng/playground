package apis

import (
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
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/ui"
)

type API interface {
	Middlewares() *ApiMiddlewares
	Api() huma.API
	RegisterRoutes()
	Router() chi.Router
	App() core.App
}

type Api struct {
	app         core.App
	api         huma.API
	router      chi.Router
	middlewares *ApiMiddlewares
}

// Middlewares implements AppApi.
func (api *Api) Middlewares() *ApiMiddlewares {
	return api.middlewares
}

var _ API = (*Api)(nil)

func (api *Api) RegisterRoutes() {
	bindMiddlewares(api)
	bindApis(api.Api(), api)
}

func (api *Api) Api() huma.API {
	if api.api == nil {
		panic("api not initialized for api")
	}
	return api.api
}

func (api *Api) Router() chi.Router {
	if api.router == nil {
		panic("router not initialized for api")
	}
	return api.router
}

func (api *Api) App() core.App {
	if api.app == nil {
		panic("app not initialized for api")
	}
	return api.app
}

func NewAppApiWithRouter(app core.App) *Api {
	router := NewRouter(app)
	api := newApiGroup(router)
	return &Api{
		app:         app,
		api:         api,
		router:      router,
		middlewares: newApiMiddlewares(api, app),
	}
}
func NewAppApi(app core.App, router chi.Router, api huma.API) *Api {
	return &Api{
		app:         app,
		api:         api,
		router:      router,
		middlewares: newApiMiddlewares(api, app),
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
		Schema: httplog.SchemaOTEL.Concise(true),

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

func newApiGroup(r chi.Router) huma.API {
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

type HumaMiddlewareFunc func(ctx huma.Context, next func(huma.Context))

type ApiMiddlewares struct {
	// customer middlewares
	SelectCustomerFromUser HumaMiddlewareFunc
	SelectCustomerFromTeam HumaMiddlewareFunc
	// team info middlewares
	TeamInfoFromParam           HumaMiddlewareFunc
	TeamInfoFromTeamSlug        HumaMiddlewareFunc
	TeamInfoFromUserAndMemberID HumaMiddlewareFunc
	TeamInfoFromTask            HumaMiddlewareFunc
	TeamInfoFromTaskProject     HumaMiddlewareFunc
	// check middlewares
	MemberIdBelongsToUser   HumaMiddlewareFunc
	TeamCanDelete           HumaMiddlewareFunc
	EmailVerified           HumaMiddlewareFunc
	TeamRequiredOwnerMember HumaMiddlewareFunc
	TeamRequiredAnyMember   HumaMiddlewareFunc
	// auth middlewares
	Auth        HumaMiddlewareFunc
	RequireAuth HumaMiddlewareFunc
	// common middlewares
	Recoverer HumaMiddlewareFunc
}

func newApiMiddlewares(api huma.API, app core.App) *ApiMiddlewares {
	return &ApiMiddlewares{
		// customer middlewares
		SelectCustomerFromUser: humamiddleware.SelectCustomerFromUser(api, app),
		SelectCustomerFromTeam: humamiddleware.SelectCustomerFromTeam(api, app),
		// team info middlewares
		TeamInfoFromParam:           humamiddleware.TeamInfoFromParam(api, app),
		TeamInfoFromTeamSlug:        humamiddleware.TeamInfoFromTeamSlug(api, app),
		TeamInfoFromUserAndMemberID: humamiddleware.TeamInfoFromUserAndMemberID(api, app),
		TeamInfoFromTask:            humamiddleware.TeamInfoFromTask(api, app),
		TeamInfoFromTaskProject:     humamiddleware.TeamInfoFromTaskProject(api, app),
		// check middlewares
		MemberIdBelongsToUser:   humamiddleware.MemberIdBelongsToUser(api, app),
		TeamCanDelete:           humamiddleware.TeamCanDelete(api, app),
		EmailVerified:           humamiddleware.HumaEmailVerifiedMiddleware(api, app),
		TeamRequiredOwnerMember: humamiddleware.RequireTeamMemberRolesMiddleware(api, models.TeamMemberRoleOwner),
		TeamRequiredAnyMember:   humamiddleware.RequireTeamMemberRolesMiddleware(api),
		// auth middlewares
		Auth:        humamiddleware.HumaAuthMiddleware(api, app),
		RequireAuth: humamiddleware.HumaRequireAuthMiddleware(api, app),
		// common middlewares
		Recoverer: humamiddleware.HumaChiMiddleware(middleware.RecovererMiddleware(app)),
	}
}
