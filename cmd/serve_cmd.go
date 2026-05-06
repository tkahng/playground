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
	database "github.com/tkahng/playground/internal/database"
	appOtel "github.com/tkahng/playground/internal/tools/otel"
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
		},
	}
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
	return serveCmd
}

func migrate(dbUrl string) error {
	mConfig := database.MigratorConfig{
		DatabaseURL: dbUrl,
	}
	migrator := database.NewMigrator(&mConfig)
	return migrator.CreateAndMigrate()
}

func Run2() error {
	firstCtx, firstCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	defer firstCancel()

	otelShutdown, err := appOtel.Setup(firstCtx, "playground", "1.0.0")
	if err != nil {
		slog.Warn("otel setup failed, continuing without telemetry", slog.Any("error", err))
	} else {
		defer func() {
			shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := otelShutdown(shutCtx); err != nil {
				slog.Warn("otel shutdown error", slog.Any("error", err))
			}
		}()
	}

	opts := conf.AppConfigGetter()
	// migrate database
	if err := migrate(opts.Db.GetDatabaseURL()); err != nil {
		return err
	}
	app := core.NewApp(opts)
	appApi := apis.NewAppApiWithRouter(app)
	appApi.RegisterRoutes()
	if port == 0 {
		port = 8080
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		Handler: appApi.Router(),
	}
	serverShutdownErr := make(chan error, 1)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)

		quitSignal := <-quit
		signal.Stop(quit)
		slog.Info(fmt.Sprintf("quit signal: %q received. starting graceful shutdown\n", quitSignal.String()), slog.String("signal", quitSignal.String()))
		firstCancel()

		ctx, cancel := context.WithTimeout(firstCtx, 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			serverShutdownErr <- err
			return
		}
		appApi.App().Close()
		serverShutdownErr <- nil
	}()

	app.RunBackgroundProcesses(firstCtx)

	slog.Info("starting server", slog.Int("port", port))

	if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-serverShutdownErr; err != nil {
		return err
	}

	return nil
}
