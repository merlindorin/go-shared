package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type HTTPServer struct {
	Host              string        `name:"http-host" help:"Host to bind the server to" default:"0.0.0.0"`
	Port              int           `name:"http-port" help:"Port to bind the server to" default:"8080"`
	ReadHeaderTimeout time.Duration `name:"http-read-header-timeout" help:"Max duration for reading the request header" default:"5s"`
	ReadTimeout       time.Duration `name:"http-read-timeout" help:"Max duration for reading the entire request" default:"30s"`
	WriteTimeout      time.Duration `name:"http-write-timeout" help:"Max duration for writing the response" default:"60s"`
	IdleTimeout       time.Duration `name:"http-idle-timeout" help:"Max duration to wait for the next request" default:"120s"`
	GracefulPeriod    time.Duration `name:"http-graceful-period" help:"Period to wait for graceful shutdown" default:"5s"`
}

func (srv *HTTPServer) Addr() string {
	return fmt.Sprintf("%s:%d", srv.Host, srv.Port)
}

func (srv *HTTPServer) Start(ctx context.Context, logger *zap.Logger, router http.Handler) func() error {
	return func() error {
		logger.Info("starting HTTP server")

		server := &http.Server{
			Addr:              srv.Addr(),
			Handler:           router,
			ReadTimeout:       srv.ReadTimeout,
			ReadHeaderTimeout: srv.ReadHeaderTimeout,
			WriteTimeout:      srv.WriteTimeout,
			IdleTimeout:       srv.IdleTimeout,
		}

		errCh := make(chan error, 1)

		go func() {
			logger.Info(
				"HTTP server listening",
				zap.String("address", srv.Addr()),
				zap.Duration("read-timeout", srv.ReadTimeout),
				zap.Duration("read-header-timeout", srv.ReadHeaderTimeout),
				zap.Duration("write-timeout", srv.WriteTimeout),
				zap.Duration("idle-timeout", srv.IdleTimeout),
				zap.Duration("graceful-period", srv.GracefulPeriod),
			)

			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errCh <- err
			}

			close(errCh)
		}()

		select {
		case err := <-errCh:
			return err
		case <-ctx.Done():
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), srv.GracefulPeriod)
		defer cancel()

		logger.Info("shutting down HTTP server")

		return server.Shutdown(shutdownCtx)
	}
}
