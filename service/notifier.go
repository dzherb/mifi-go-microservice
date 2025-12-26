package service

import (
	"encoding/json"
	"log/slog"
	"time"
)

type Notifier struct {
	log *slog.Logger
}

func NewNotifier(log *slog.Logger) *Notifier {
	return &Notifier{
		log: log,
	}
}

func (n *Notifier) Send(msg string, extra map[string]any) {
	if extra == nil {
		extra = make(map[string]any)
	}

	extra["message"] = msg
	extra["timestamp"] = time.Now().Unix()

	notificationData, err := json.Marshal(extra)
	if err != nil {
		n.log.Error(
			"error marshaling notification",
			slog.String("error", err.Error()),
		)

		return
	}

	// imitate sending the notification
	time.Sleep(100 * time.Millisecond)

	n.log.Info(
		"notification sent",
		slog.String("payload", string(notificationData)),
	)
}
