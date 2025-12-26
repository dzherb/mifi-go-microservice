package response

import "time"

type Ping struct {
	ServerTime int64 `json:"server_time"`
}

func NewPing() *Ping {
	return &Ping{
		ServerTime: time.Now().Unix(),
	}
}
