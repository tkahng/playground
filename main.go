package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tkahng/authgo/cmd"
	"github.com/tkahng/authgo/internal/apis"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/core"
)

// Options for the CLI.
type Options struct {
	Port int `help:"Port to listen on" short:"p" default:"8080"`
}

// GreetingOutput represents the greeting operation response.
type GreetingOutput struct {
	Body struct {
		Message string `json:"message" example:"Hello, world!" doc:"Greeting message"`
	}
}

func main() {
	// Create a CLI app which takes a port option.
	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		ctx := context.Background()
		// Create a new router & API
		config := apis.InitApiConfig()
		// config.DocsPath = ""
		// r := http.newser
		r := chi.NewMux()
		r.Use(middleware.Logger)
		r.Use(middleware.Recoverer)
		api := humachi.New(r, config)
		grp := huma.NewGroup(api, "/api")
		// Register GET /greeting/{name}
		opts := conf.AppConfigGetter()
		app := core.InitBaseApp(ctx, opts)
		apis.AddRoutes(grp, app)

		server := http.Server{
			Addr:    fmt.Sprintf(":%d", options.Port),
			Handler: r,
		}
		// Tell the CLI how to start your server.
		hooks.OnStart(func() {
			fmt.Printf("Starting server on port %d...\n", options.Port)
			server.ListenAndServe()
		})
		// Tell the CLI how to stop your server.
		hooks.OnStop(func() {
			// Give the server 5 seconds to gracefully shut down, then give up.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			server.Shutdown(ctx)
		})
	})
	cli.Root().AddCommand(cmd.NewMigrateCmd(), cmd.NewSeedCmd())
	// Run the CLI. When passed no commands, it starts the server.
	cli.Run()
}
