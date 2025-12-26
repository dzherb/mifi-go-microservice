package server

import (
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/dzherb/mifi-go-microservice/server/handler"
	"github.com/dzherb/mifi-go-microservice/server/middleware"
	"github.com/dzherb/mifi-go-microservice/service"
)

type APIConfig struct {
	MaxRequestsPerSecond int
	MaxBurst             int
}

func RootHandler(
	log *slog.Logger,
	userService *service.UserService,
	notifier *service.Notifier,
	cfg *APIConfig,
) http.Handler {
	r := mux.NewRouter()

	r.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)

	api := r.PathPrefix("/api").Subrouter()

	pingHandler := handler.NewPingHandler(log)
	api.Handle(
		"/ping",
		http.HandlerFunc(pingHandler.Ping),
	).Methods(http.MethodGet)

	userHandler := handler.NewUserHandler(log, userService, notifier)
	api.Handle(
		"/users/{id}",
		http.HandlerFunc(userHandler.Get),
	).Methods(http.MethodGet)
	api.Handle(
		"/users",
		http.HandlerFunc(userHandler.GetAll),
	).Methods(http.MethodGet)
	api.Handle(
		"/users",
		http.HandlerFunc(userHandler.Create),
	).Methods(http.MethodPost)

	api.Use(middleware.CollectRequestsMetrics)
	api.Use(middleware.RateLimitMiddleware(
		float64(cfg.MaxRequestsPerSecond),
		cfg.MaxBurst,
	),
	)

	return r
}

type Config struct {
	Host              string
	Port              int
	ReadHeaderTimeout time.Duration
}

func New(h http.Handler, cfg Config) *http.Server {
	return &http.Server{
		Addr:              net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Handler:           h,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}
}
