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

type Ps2ServiceConfig struct {
	HttpClientTimeout time.Duration `yaml:"http_client_timeout" env:"PS2_SERVICE_HTTP_CLIENT_TIMEOUT" env-default:"8s"`
}

type BotConfig struct {
	CommandHandlerTimeout time.Duration `yaml:"command_handler_timeout" env:"BOT_COMMAND_HANDLER_TIMEOUT" env-default:"20s"`
	DiscordToken          string        `yaml:"token" env:"BOT_DISCORD_TOKEN" env-required:"true"`
}

type StorageConfig struct {
	Path string `yaml:"path" env:"STORAGE_PATH" env-required:"true"`
}

type Config struct {
	Logger     LoggerConfig     `yaml:"logger"`
	Storage    StorageConfig    `yaml:"storage"`
	Ps2Service Ps2ServiceConfig `yaml:"ps2_service"`
	Bot        BotConfig        `yaml:"bot"`
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
