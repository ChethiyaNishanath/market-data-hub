package binance

type DepthUpdateEvent struct {
	EventTime          int        `json:"E"`
	FirstUpdateEventID int        `json:"U"`
	FinalUpdateEventID int        `json:"u"`
	Symbol             string     `json:"s"`
	EventType          string     `json:"e"`
	BidsToUpdated      [][]string `json:"b"`
	AsksToUpdated      [][]string `json:"a"`
}

type DepthStreamEvent struct {
	EventType          string     `json:"e"`
	EventTime          int        `json:"E"`
	Symbol             string     `json:"s"`
	FirstUpdateEventID int        `json:"U"`
	FinalUpdateEventID int        `json:"u"`
	BidsToUpdated      [][]string `json:"b"`
	AsksToUpdated      [][]string `json:"a"`
}

type OrderBookStreamEvent struct {
	Action       string     `json:"action"`
	LastUpdateID int        `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}
