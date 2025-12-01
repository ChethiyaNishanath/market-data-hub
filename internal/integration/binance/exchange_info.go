package binance

type ExchangeInfo struct {
	TimeZone        string   `json:"timeZone"`
	ServerTime      int64    `json:"serverTime"`
	RateLimits      any      `json:"rateLimits"`
	ExchangeFilters any      `json:"exchangeFilters"`
	Symbols         []Symbol `json:"symbols"`
	Sors            any      `json:"sors"`
}

type Symbol struct {
	Symbol                          string   `json:"symbol"`
	Status                          string   `json:"status"`
	BaseAsset                       string   `json:"baseAsset"`
	BaseAssetPrecision              int      `json:"baseAssetPrecision"`
	QuoteAsset                      string   `json:"quoteAsset"`
	QuotePrecision                  int      `json:"quotePrecision"`
	QuoteAssetPrecision             int      `json:"quoteAssetPrecision"`
	BaseCommissionPrecision         int      `json:"baseCommissionPrecision"`
	QuoteCommissionPrecision        int      `json:"quoteCommissionPrecision"`
	OrderTypes                      []string `json:"orderTypes"`
	IcebergAllowed                  bool     `json:"icebergAllowed"`
	OcoAllowed                      bool     `json:"ocoAllowed"`
	OtoAllowed                      bool     `json:"otoAllowed"`
	QuoteOrderQtyMarketAllowed      bool     `json:"quoteOrderQtyMarketAllowed"`
	AllowTrailingStop               bool     `json:"allowTrailingStop"`
	CancelReplaceAllowed            bool     `json:"cancelReplaceAllowed"`
	AmendAllowed                    bool     `json:"amendAllowed"`
	PegInstructionsAllowed          bool     `json:"pegInstructionsAllowed"`
	IsSpotTradingAllowed            bool     `json:"isSpotTradingAllowed"`
	IsMarginTradingAllowed          bool     `json:"isMarginTradingAllowed"`
	Filters                         any      `json:"filters"`
	Permissions                     any      `json:"permissions"`
	PermissionSets                  any      `json:"permissionSets"`
	DefaultSelfTradePreventionMode  string   `json:"defaultSelfTradePreventionMode"`
	AllowedSelfTradePreventionModes []string `json:"allowedSelfTradePreventionModes"`
}
