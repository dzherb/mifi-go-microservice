package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/dzherb/mifi-go-microservice/server/response"
)

type Ping struct {
	ServerTime int64 `json:"server_time"`
}

func NewPing() *Ping {
	return &Ping{
		ServerTime: time.Now().Unix(),
	}
}

type PingHandler struct {
	log *slog.Logger
}

func NewPingHandler(log *slog.Logger) *PingHandler {
	return &PingHandler{log: log}
}

func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	response.Write(w, h.log, NewPing(), http.StatusOK)
}
