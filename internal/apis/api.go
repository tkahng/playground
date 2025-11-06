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
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/httprate"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	appHttp "github.com/tkahng/playground/internal/tools/http"
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
		middlewares: newApiMiddlewares(app),
	}
}
func NewAppApi(app core.App, router chi.Router, api huma.API) *Api {
	return &Api{
		app:         app,
		api:         api,
		router:      router,
		middlewares: newApiMiddlewares(app),
	}
}

func AddBaseMiddlewares(app core.App, r chi.Router, mw ...func(http.Handler) http.Handler) {
	r.Use(chimiddleware.RequestID)
	r.Use()
	r.Use(httplog.RequestLogger(app.Logger(), &httplog.Options{
		Level: slog.LevelInfo,

		// Set log output to Elastic Common Schema (ECS) format.
		Schema: httplog.SchemaOTEL.Concise(true),

		// RecoverPanics recovers from panics occurring in the underlying HTTP handlers
		// and middlewares. It returns HTTP 500 unless response status was already set.
		//
		// NOTE: Panics are logged as errors automatically, regardless of this setting.
		RecoverPanics: true,
	}))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "PUT", "POST", "DELETE", "HEAD", "OPTION"},
		AllowedHeaders: []string{
			"User-Agent",
			"Content-Type",
			"Accept",
			"Accept-Encoding",
			"Accept-Language",
			"Cache-Control",
			"Connection",
			"DNT",
			"Host",
			"Origin",
			"Pragma",
			"Referer",
			"X-Client-IP",
			"X-Forwarded-For",
			"X-Forwarded",
			"Forwarded-For",
			"Forwarded",
			"CF-Connecting-IP",
			"Fastly-Client-Ip",
			"True-Client-Ip",
			"X-Real-IP",
			"X-Cluster-Client-IP",
		},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
	}))
	r.Use(httprate.LimitByIP(100, 1*time.Minute))
	r.Use(mw...)
}

func NewRouter(app core.App) *chi.Mux {
	r := chi.NewMux()
	r.Use(chimiddleware.RequestID)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			requestId := chimiddleware.GetReqID(rawCtx)
			if requestId == "" {
				_ = appHttp.WriteErr(w, r, http.StatusUnauthorized, "unauthorized. user info not found", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	})
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
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
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

func newApiMiddlewares(app core.App) *ApiMiddlewares {
	return &ApiMiddlewares{
		// customer middlewares
		SelectCustomerFromUser: humamiddleware.HumaChiMiddleware(middleware.SelectCustomerFromUser(app)),
		SelectCustomerFromTeam: humamiddleware.HumaChiMiddleware(middleware.SelectCustomerFromTeam(app)),
		// team info middlewares
		TeamInfoFromParam:           humamiddleware.HumaChiMiddleware(middleware.TeamInfoFromParam(app)),
		TeamInfoFromTeamSlug:        humamiddleware.HumaChiMiddleware(middleware.TeamInfoFromTeamSlug(app)),
		TeamInfoFromUserAndMemberID: humamiddleware.HumaChiMiddleware(middleware.TeamInfoFromUserAndMemberID(app)),
		TeamInfoFromTask:            humamiddleware.HumaChiMiddleware(middleware.TeamInfoFromTask(app)),
		TeamInfoFromTaskProject:     humamiddleware.HumaChiMiddleware(middleware.TeamInfoFromTaskProject(app)),
		// check middlewares
		MemberIdBelongsToUser:   humamiddleware.HumaChiMiddleware(middleware.MemberIdBelongsToUser(app)),
		TeamCanDelete:           humamiddleware.HumaChiMiddleware(middleware.TeamCanDelete(app)),
		EmailVerified:           humamiddleware.HumaChiMiddleware(middleware.HttpEmailVerifiedMiddleware()),
		TeamRequiredOwnerMember: humamiddleware.HumaChiMiddleware(middleware.RequireTeamMemberRolesMiddleware(models.TeamMemberRoleOwner)),
		TeamRequiredAnyMember:   humamiddleware.HumaChiMiddleware(middleware.RequireTeamMemberRolesMiddleware()),
		// auth middlewares
		Auth:        humamiddleware.HumaChiMiddleware(middleware.HttpAuthMiddleware(app)),
		RequireAuth: humamiddleware.HumaChiMiddleware(middleware.HttpRequireAuthMiddleware()),
		// common middlewares
		Recoverer: humamiddleware.HumaChiMiddleware(middleware.RecovererMiddleware()),
	}
}
