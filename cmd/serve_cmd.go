package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

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
			if err := run(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
			// serve(ctx)
		},
	}
	serveCmd.Flags().IntP("port", "p", 8080, "Port to listen on")
	return serveCmd
}

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	opts := conf.AppConfigGetter()
	app := core.InitBaseApp(ctx, opts)
	srv, api := NewServer()
	apis.AddRoutes(api, app)
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: srv,
	}
	go func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()
	return nil
}

func NewServer() (http.Handler, huma.API) {
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
	// grp := huma.NewGroup(api, "/api")
	return r, api
}
