package config

type Config struct {
	Server       Server             `mapstructure:"server"`
	Integrations IntegrationsConfig `mapstructure:"integrations"`
	Logging      Logging            `mapstructure:"logging"`
}

type Logging struct {
	Level string `mapstructure:"level"`
}

type Server struct {
	Port            string `mapstructure:"port"`
	ShutdownTimeout string `mapstructure:"shutdownTimeout"`
}

type IntegrationsConfig struct {
	Binance BinanceConfig `mapstructure:"binance"`
}

type BinanceConfig struct {
	WsStreamUrl   string `mapstructure:"wsStreamUrl"`
	RestApiUrlV3  string `mapstructure:"restApiUrlV3"`
	Subscriptions string `mapstructure:"subscriptions"`
}
