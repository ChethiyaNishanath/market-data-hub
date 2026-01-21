package binance

const (
	OrderBookReset  = "orderbookReset"
	OrderBookUpdate = "depthUpdate"
)
//FEEDBACK : no need to have multiple files to define different messages. move it to one called message.go or model.go