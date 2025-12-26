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

	"github.com/dzherb/mifi-go-microservice/logger"
	"github.com/dzherb/mifi-go-microservice/server"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

	log := logger.New(slog.LevelInfo)

	srv := server.New(
		server.APIHandler(log),
		server.Config{
			Host:              "localhost",
			Port:              8080,
			ReadHeaderTimeout: 10 * time.Second,
		},
	)

	serverDone := make(chan struct{})

	go func() {
		log.Info(
			"starting api server",
			slog.String("address", srv.Addr),
		)

		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(
				"server unexpectedly stopped",
				slog.String("error", err.Error()),
			)
		}

		serverDone <- struct{}{}
	}()

	<-ctx.Done()

	log.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	err := srv.Shutdown(shutdownCtx)
	if err != nil {
		log.Error(
			"server shutdown error",
			slog.String("error", err.Error()),
		)
	}

	select {
	case <-shutdownCtx.Done():
	case <-serverDone:
	}
}
