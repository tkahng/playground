package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tkahng/authgo/internal/apis"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/core"
	"golang.org/x/sync/errgroup"
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
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
	return serveCmd
}

func run(ctx context.Context) error {
	ctx = context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)
	opts := conf.AppConfigGetter()
	app := core.NewBaseApp(ctx, opts)
	appApi := apis.NewApi(app)
	startEvent := apis.NewStartEvent(opts)
	appApi.BindApi(startEvent.Api)
	if port == 0 {
		port = 8080
	}

	// Run HTTP server
	g.Go(func() error {
		slog.Info("Starting HTTP server", "addr", startEvent.Server.Addr)
		if err := startEvent.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("http server error: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		slog.Info("Starting poller")
		return app.JobManager().Run(ctx)
	})
	err := g.Wait()
	if err != nil {
		return fmt.Errorf("error running server: %w", err)
	}
	// Gracefully shutdown HTTP server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = startEvent.Server.Shutdown(shutdownCtx)

	return err
	// go func() {
	// 	log.Printf("listening on %s\n", httpServer.Addr)
	// 	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	// 		fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
	// 	}
	// }()
	// var wg sync.WaitGroup
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	<-ctx.Done()
	// 	shutdownCtx := context.Background()
	// 	shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
	// 	defer cancel()
	// 	if err := httpServer.Shutdown(shutdownCtx); err != nil {
	// 		fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
	// 	}
	// }()
	// wg.Wait()
	// return nil
}
