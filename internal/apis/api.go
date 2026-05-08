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
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/ui"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type API interface {
	Middlewares() *humamiddleware.ApiMiddlewares
	Api() huma.API
	RegisterRoutes()
	Router() chi.Router
	App() core.App
}

type Api struct {
	app         core.App
	api         huma.API
	router      chi.Router
	middlewares *humamiddleware.ApiMiddlewares
}

// Middlewares implements AppApi.
func (api *Api) Middlewares() *humamiddleware.ApiMiddlewares {
	return api.middlewares
}

var _ API = (*Api)(nil)

func (api *Api) RegisterRoutes() {
	bindMiddlewares(api)
	bindApis(api)
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
		middlewares: humamiddleware.NewApiMiddlewares(app),
	}
}
func NewAppApi(app core.App, router chi.Router, api huma.API) *Api {
	return &Api{
		app:         app,
		api:         api,
		router:      router,
		middlewares: humamiddleware.NewApiMiddlewares(app),
	}
}

func AddBaseMiddlewares(app core.App, r chi.Router, mw ...func(http.Handler) http.Handler) {
	r.Use(otelhttp.NewMiddleware("http.server"))
	r.Use(middleware.InitContextAttrsMiddleware)
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.SetRequestIDAttrsMiddleware)
	r.Use(httplog.RequestLogger(app.Logger(), &httplog.Options{
		Level:  slog.LevelInfo,
		Schema: httplog.SchemaOTEL.Concise(true),
		Skip: func(req *http.Request, respStatus int) bool {
			return req.URL.Path == "/actuator/health"
		},
		RecoverPanics: true,
	}))
	r.Use(cors.Handler(cors.Options{
		AllowOriginFunc: func(r *http.Request, origin string) bool { return true },
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

	AddBaseMiddlewares(app, r)
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
