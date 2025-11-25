package ws

import "encoding/json"

type WSRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type WSMessage struct {
	Method  string `json:"method,omitempty"`
	Success bool   `json:"success,omitempty"`
	Error   string `json:"error,omitempty"`
	Topic   string `json:"topic,omitempty"`
	Data    any    `json:"data,omitempty"`
}
