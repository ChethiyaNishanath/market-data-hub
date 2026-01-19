package binance

import "encoding/json"

type DepthUpdateMessage struct {
	EventTime          int        `json:"E"`
	FirstUpdateEventID int        `json:"U"`
	FinalUpdateEventID int        `json:"u"`
	Symbol             string     `json:"s"`
	EventType          string     `json:"e"`
	BidsToUpdated      [][]string `json:"b"`
	AsksToUpdated      [][]string `json:"a"`
}

type OrderBookDepthUpdateStreamMessage struct {
	Action       string     `json:"action"`
	LastUpdateID int        `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

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
