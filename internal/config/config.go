package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	TextHandler = "text"
	JSONHandler = "json"
)

const (
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warn"
	ErrorLevel = "error"
)

type LoggerConfig struct {
	Level       string `yaml:"level" env:"LOGGER_LEVEL" env-default:"info"`
	HandlerType string `yaml:"handler_type" env:"LOGGER_HANDLER_TYPE" env-default:"text"`
}

type BotConfig struct {
	CommandHandlerTimeout  time.Duration `yaml:"command_handler_timeout" env:"BOT_COMMAND_HANDLER_TIMEOUT" env-default:"20s"`
	Ps2EventHandlerTimeout time.Duration `yaml:"ps2_event_handler_timeout" env:"BOT_PS2_EVENT_HANDLER_TIMEOUT" env-default:"2m"`
	DiscordToken           string        `yaml:"token" env:"BOT_DISCORD_TOKEN" env-required:"true"`
	HttpClientTimeout      time.Duration `yaml:"http_client_timeout" env:"BOT_HTTP_CLIENT_TIMEOUT" env-default:"8s"`
	CensusServiceId        string        `yaml:"census_service_id" env:"BOT_CENSUS_SERVICE_ID" env-required:"true"`
}

type StorageConfig struct {
	Path string `yaml:"path" env:"STORAGE_PATH" env-required:"true"`
}

type Config struct {
	Logger  LoggerConfig  `yaml:"logger"`
	Storage StorageConfig `yaml:"storage"`
	Bot     BotConfig     `yaml:"bot"`
}

func MustLoad(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file not found: %s", configPath)
	}
	cfg := &Config{}
	if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}
	return cfg
}
