package config

import (
	"log/slog"
	"os"
)

type Config struct {
	Port            string
	ShutdownTimeout string
	DBDsn           string
	Binance         *BinanceConfig
}

type BinanceConfig struct {
	WsStreamUrl       string
	WsRestApiUrlV3    string
	SubscribedSymbols string
}

var globalConfig *Config

func Init() {
	globalConfig = load()
}

func GetConfig() *Config {
	return globalConfig
}

func load() *Config {
	slog.Debug("Loading environment variables")
	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		ShutdownTimeout: getEnv("SHUTDOWN_TIMEOUT", "10"),

		DBDsn: getEnv("DB_DSN", "postgres://postgres:Test1234@localhost:5432/usermanagementdb?sslmode=disable"),
		Binance: &BinanceConfig{
			WsStreamUrl:       getEnv("BINACE_WS_STREAM_URL", "wss://stream.binance.com:9443/ws"),
			WsRestApiUrlV3:    getEnv("BINANCE_REST_URL_V3", "https://api.binance.com/api/v3"),
			SubscribedSymbols: getEnv("BINANCE_SUBSCRIBED_SYMBOLS", "BTCUSDT,BNBBTC"),
		},
	}
	return cfg
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
