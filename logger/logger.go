package logger

import (
	"log/slog"
	"os"
)

func New(levelName string) *slog.Logger {
	var lvl slog.Level

	err := lvl.UnmarshalText([]byte(levelName))
	if err != nil {
		panic("parse log level: " + err.Error())
	}

	return slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: lvl,
			},
		),
	)
}
