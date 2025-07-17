package cmd

import (
	"context"
	"errors"
	"fmt"
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
	firstCtx, firstCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
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

		ctx, cancel := context.WithTimeout(firstCtx, 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			serverShutdownErr <- err
			return
		}

		serverShutdownErr <- nil
	}()

	app.RunBackgroundProcesses(firstCtx)

	fmt.Printf("server running on port %d", app.Config().Options.Port)

	if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-serverShutdownErr; err != nil {
		return err
	}

	return nil

}
