package exchange

import "time"

type Client struct {
	URL          string
	PingInterval time.Duration
}

func New(url string) Client {
	return Client{
		URL:          url,
		PingInterval: 20 * time.Second,
	}
}
