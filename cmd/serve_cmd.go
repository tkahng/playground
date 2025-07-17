package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/core"
)

var port int

func NewServeCmd() *cobra.Command {
	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP server",
		Long:  `Starts the HTTP server on a specified port`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := Run2(); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
			// serve(ctx)
		},
	}
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
	return serveCmd
}

func Run2() error {
	ctx, firstCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	defer firstCancel()
	opts := conf.AppConfigGetter()
	app := core.BootstrappedApp(opts)
	appApi := apis.NewApi(app)
	srv, api := apis.NewServer()
	apis.AddRoutes(api, appApi)
	if port == 0 {
		port = 8080
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: srv,
	}
	serverShutdownErr := make(chan error, 1)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

		quitSignal := <-quit
		signal.Stop(quit)

		fmt.Printf("quit signal: %q received. starting graceful shutdown\n", quitSignal.String())

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			serverShutdownErr <- err
			return
		}

		serverShutdownErr <- nil
	}()

	go func() {
		slog.Info("Starting poller")
		if err := app.JobManager().Run(context.Background()); err != nil {
			slog.ErrorContext(
				ctx,
				"error starting poller",
				slog.Any("error", err),
			)
		}
	}()

	go func() {
		slog.Info("Starting sse manager")
		app.SseManager().Run(context.Background())
	}()

	fmt.Printf("server running on port %d", app.Config().Options.Port)

	if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-serverShutdownErr; err != nil {
		return err
	}

	return nil

}

// func Run() error {
// 	ctx, firstCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
// 	defer firstCancel()
// 	g, ctx := errgroup.WithContext(ctx)
// 	opts := conf.AppConfigGetter()
// 	app := core.BootstrappedApp(opts)
// 	appApi := apis.NewApi(app)
// 	srv, api := apis.NewServer()
// 	apis.AddRoutes(api, appApi)
// 	if port == 0 {
// 		port = 8080
// 	}

// 	httpServer := &http.Server{
// 		Addr:    fmt.Sprintf(":%d", port),
// 		Handler: srv,
// 	}

// 	g.Go(func() error {
// 		slog.Info("Starting notifier listener")
// 		if err := app.Notifier().Start(ctx); err != nil {
// 			slog.ErrorContext(
// 				ctx,
// 				"error starting notifier listener",
// 				slog.Any("error", err),
// 			)
// 			return err
// 		}
// 		return nil
// 	})
// 	g.Go(func() error {
// 		slog.Info("Starting poller")
// 		if err := app.JobManager().Run(ctx); err != nil {
// 			slog.ErrorContext(
// 				ctx,
// 				"error starting poller",
// 				slog.Any("error", err),
// 			)
// 			return err
// 		}
// 		return nil
// 	})
// 	// Run HTTP server
// 	g.Go(func() error {
// 		slog.Info("Starting HTTP server", "addr", httpServer.Addr)
// 		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 			return fmt.Errorf("http server error: %w", err)
// 		}
// 		return nil
// 	})
// 	err := g.Wait()
// 	if err != nil {
// 		return err
// 	}
// 	// Gracefully shutdown HTTP server
// 	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer shutdownCancel()
// 	err = httpServer.Shutdown(shutdownCtx)

// 	return err
// 	// go func() {
// 	// 	log.Printf("listening on %s\n", httpServer.Addr)
// 	// 	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 	// 		fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
// 	// 	}
// 	// }()
// 	// var wg sync.WaitGroup
// 	// wg.Add(1)
// 	// go func() {
// 	// 	defer wg.Done()
// 	// 	<-ctx.Done()
// 	// 	shutdownCtx := context.Background()
// 	// 	shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
// 	// 	defer cancel()
// 	// 	if err := httpServer.Shutdown(shutdownCtx); err != nil {
// 	// 		fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
// 	// 	}
// 	// }()
// 	// wg.Wait()
// 	// return nil
// }

// func RunHooks() error {

// 	baseContext, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
// 	defer cancel()
// 	// g, ctx := errgroup.WithContext(ctx)
// 	opts := conf.AppConfigGetter()
// 	app := core.BootstrappedApp(opts)
// 	appApi := apis.NewApi(app)
// 	startEvent := apis.NewStartEvent(opts)

// 	appApi.BindApi(startEvent.Api)

// 	var wg sync.WaitGroup

// 	app.Lifecycle().OnStop().Bind(&hook.Handler[*core.StopEvent]{
// 		Id: "pbGracefulShutdown",
// 		Func: func(te *core.StopEvent) error {
// 			cancel()

// 			ctx, cancel := context.WithTimeout(baseContext, 1*time.Second)
// 			defer cancel()

// 			wg.Add(1)

// 			_ = startEvent.Server.Shutdown(ctx)

// 			if te.IsRestart {
// 				// wait for execve and other handlers up to 3 seconds before exit
// 				time.AfterFunc(3*time.Second, func() {
// 					wg.Done()
// 				})
// 			} else {
// 				wg.Done()
// 			}

// 			return te.Next()
// 		},
// 		Priority: -9999,
// 	})
// 	// wait for the graceful shutdown to complete before exit
// 	defer func() {
// 		wg.Wait()

// 		if startEvent.Server != nil {
// 			_ = startEvent.Server.Close()
// 		}
// 	}()
// 	serverHookErr := app.Lifecycle().OnStart().Trigger(startEvent, func(se *core.StartEvent) error {
// 		slog.Info("Starting HTTP server", "addr", startEvent.Server.Addr)
// 		if err := startEvent.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 			return fmt.Errorf("http server error: %w", err)
// 		}
// 		return nil
// 	})

// 	if serverHookErr != nil {
// 		return serverHookErr
// 	}
// 	return nil
// 	// // Run HTTP server
// 	// g.Go(func() error {
// 	// 	slog.Info("Starting HTTP server", "addr", startEvent.Server.Addr)
// 	// 	if err := startEvent.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 	// 		return fmt.Errorf("http server error: %w", err)
// 	// 	}
// 	// 	return nil
// 	// })
// 	// g.Go(func() error {
// 	// 	slog.Info("Starting poller")
// 	// 	return app.JobManager().Run(ctx)
// 	// })
// 	// err := g.Wait()
// 	// if err != nil {
// 	// 	return fmt.Errorf("error running server: %w", err)
// 	// }
// 	// // Gracefully shutdown HTTP server
// 	// shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	// defer cancel()
// 	// _ = startEvent.Server.Shutdown(shutdownCtx)

// 	// return err
// }
