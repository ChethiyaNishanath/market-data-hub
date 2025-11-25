package marketstate

type OrderBook struct {
	LastUpdateID int        `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
	Initialized  bool       `json:"-"`
}
