package handler

import (
	"log/slog"
	"net/http"

	"github.com/dzherb/mifi-go-microservice/server/response"
)

type PingHandler struct {
	log *slog.Logger
}

func NewPingHandler(log *slog.Logger) *PingHandler {
	return &PingHandler{log: log}
}

func (h *PingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response.Write(w, h.log, response.NewPing(), http.StatusOK)
}
