package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/spf13/cobra"
	"github.com/tkahng/authgo/internal/apis"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/core"
)

var port int

func NewServeCmd() *cobra.Command {
	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP server",
		Long:  `Starts the HTTP server on a specified port`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			serve(ctx)
		},
	}
	serveCmd.Flags().IntP("port", "p", 8080, "Port to listen on")
	return serveCmd
}

func serve(ctx context.Context) {
	var api huma.API
	// ctx := context.Background()
	// Create a new router & API
	config := apis.InitApiConfig()
	// config.DocsPath = ""
	// r := http.newser
	r := chi.NewMux()
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		// MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	api = humachi.New(r, config)
	grp := huma.NewGroup(api, "/api")
	// port = fmt.Printf("Starting server on port: %d\n", port)
	// Register GET /greeting/{name}
	opts := conf.AppConfigGetter()
	app := core.InitBaseApp(ctx, opts)
	apis.AddRoutes(grp, app)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}
	// Tell the CLI how to start your server.
	fmt.Printf("Starting server on port %d...\n", port)
	if err := server.ListenAndServe(); err != nil {

	}
	// hooks.OnStart(func() {
	// })
	// Tell the CLI how to stop your server.
	// hooks.OnStop(func() {
	// 	// Give the server 5 seconds to gracefully shut down, then give up.
	// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 	defer cancel()

	// 	server.Shutdown(ctx)
	// })
}
