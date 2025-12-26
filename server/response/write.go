package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func Write(w http.ResponseWriter, log *slog.Logger, obj any, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if status == 0 {
		status = http.StatusOK
	}

	w.WriteHeader(status)

	if obj != nil {
		err := json.NewEncoder(w).Encode(obj)
		if err != nil {
			log.Error(
				"error writing response",
				slog.String("error", err.Error()),
			)
		}
	}
}
