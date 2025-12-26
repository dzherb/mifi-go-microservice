package logger

import (
	"log/slog"
	"os"
)

func New(lvl slog.Level) *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: lvl,
			},
		),
	)
}
