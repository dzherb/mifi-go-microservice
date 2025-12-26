package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/dzherb/mifi-go-microservice/logger"
	"github.com/dzherb/mifi-go-microservice/model"
	"github.com/dzherb/mifi-go-microservice/server"
	"github.com/dzherb/mifi-go-microservice/service"
	"github.com/dzherb/mifi-go-microservice/storage"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

	log := logger.New(envOrDefault("LOG_LEVEL", "info"))

	userStorage, err := storage.NewMiniIO[model.User](
		ctx,
		log,
		storage.MiniIOConfig{
			Endpoint:   envOrDefault("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey:  envOrDefault("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey:  envOrDefault("MINIO_SECRET_KEY", "minioadmin"),
			BucketName: envOrDefault("MINIO_BUCKET", "users"),
			UseSSL:     envOrDefault("MINIO_USE_SSL", "false") == "true",
		},
	)

	if err != nil {
		panic("user storage initialization: " + err.Error())
	}

	userService := service.NewUserService(log, userStorage)
	notifier := service.NewNotifier(log)

	srv := server.New(
		server.RootHandler(log, userService, notifier),
		server.Config{
			Port: intEnvOrDefault("SERVER_PORT", 8080),
			ReadHeaderTimeout: time.Duration(
				intEnvOrDefault("SERVER_READ_HEADER_TIMEOUT_IN_MS", 10000),
			) * time.Millisecond,
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

	select {
	case <-ctx.Done():
	case <-serverDone:
	}

	log.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	err = srv.Shutdown(shutdownCtx)
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

func envOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func intEnvOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		result, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}

		return result
	}

	return defaultValue
}
