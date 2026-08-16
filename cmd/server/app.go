package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

type App struct {
	httpsrv *http.Server
}

func (app *App) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		slog.InfoContext(ctx, "starting pprof http server:", slog.String("address", app.httpsrv.Addr))
		if err := app.httpsrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	errch := make(chan error, 1)
	go func() { errch <- g.Wait() }()

	select {
	case <-ctx.Done():
		stop()

		slog.InfoContext(ctx, "shutting down tephroite server")
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
		defer cancel()

		shutdownErr := app.shutdown(shutdownCtx)
		runErr := <-errch

		return errors.Join(runErr, shutdownErr)
	case err := <-errch:
		return err
	}
}

func (app *App) shutdown(ctx context.Context) error {
	slog.InfoContext(ctx, "shutting down pprof http server")
	httpErr := app.httpsrv.Shutdown(ctx)
	if httpErr != nil {
		if err := app.httpsrv.Close(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			httpErr = errors.Join(httpErr, err)
		}
	}

	return httpErr
}
