package server

import (
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/dzherb/mifi-go-microservice/server/handler"
)

func APIHandler(
	log *slog.Logger,
) http.Handler {
	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()

	api.Handle("/ping", handler.NewPingHandler(log)).Methods(http.MethodGet)

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
