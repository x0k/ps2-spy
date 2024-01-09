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

type HttpClientConfig struct {
	Timeout     time.Duration `yaml:"timeout" env:"HTTP_CLIENT_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"HTTP_CLIENT_IDLE_TIMEOUT" env-default:"10s"`
}

type BotConfig struct {
	CommandHandlerTimeout time.Duration `yaml:"command_handler_timeout" env:"COMMAND_HANDLER_TIMEOUT" env-default:"20s"`
	DiscordToken          string        `yaml:"token" env:"DISCORD_TOKEN" env-required:"true"`
}

type StorageConfig struct {
	Path string `yaml:"path" env:"STORAGE_PATH" env-required:"true"`
}

type Config struct {
	LoggerConfig LoggerConfig     `yaml:"logger"`
	Storage      StorageConfig    `yaml:"storage"`
	HttpClient   HttpClientConfig `yaml:"http_client"`
	Bot          BotConfig        `yaml:"bot"`
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
