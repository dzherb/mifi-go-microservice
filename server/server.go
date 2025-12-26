package server

import (
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/dzherb/mifi-go-microservice/server/handler"
	"github.com/dzherb/mifi-go-microservice/service"
)

func RootHandler(
	log *slog.Logger,
	userService *service.UserService,
	notifier *service.Notifier,
) http.Handler {
	r := mux.NewRouter()

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

	return api
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
