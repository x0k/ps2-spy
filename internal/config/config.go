package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type HttpClientConfig struct {
	Timeout     time.Duration `yaml:"timeout" env:"HTTP_CLIENT_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"HTTP_CLIENT_IDLE_TIMEOUT" env-default:"10s"`
}

type DiscordConfig struct {
	Token string `yaml:"token" env:"DISCORD_TOKEN" env-required:"true"`
}

type Config struct {
	Env         string           `yaml:"env" env:"ENV" env-required:"true"`
	StoragePath string           `yaml:"storage_path" env:"STORAGE_PATH" env-required:"true"`
	HttpClient  HttpClientConfig `yaml:"http_client"`
	Discord     DiscordConfig    `yaml:"discord"`
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
