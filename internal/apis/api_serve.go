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
	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/httprate"

	"github.com/go-chi/cors"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/tools/logger"
	"github.com/tkahng/playground/ui"
)

func NewServer(opt *conf.EnvConfig) (http.Handler, huma.API) {
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

	// config.DocsPath = ""

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
	r.Use(httplog.RequestLogger(logger.GetDefaultLogger(), &httplog.Options{
		// Level defines the verbosity of the request logs:
		// slog.LevelDebug - log all responses (incl. OPTIONS)
		// slog.LevelInfo  - log responses (excl. OPTIONS)
		// slog.LevelWarn  - log 4xx and 5xx responses only (except for 429)
		// slog.LevelError - log 5xx responses only
		Level: slog.LevelInfo,

		// Set log output to Elastic Common Schema (ECS) format.
		Schema: logger.GetDefaultFormat(&opt.AppConfig),

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

	api = humachi.New(r, config)

	// swagger
	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <meta name="description" content="SwaggerUI" />
  <title>SwaggerUI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin></script>
<script>
  window.onload = () => {
    window.ui = SwaggerUIBundle({
      url: '/openapi.json',
      dom_id: '#swagger-ui',
    });
  };
</script>
</body>
</html>`))
	})

	grp := huma.NewGroup(api, "/api")
	return r, grp
}
