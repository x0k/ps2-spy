package app

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	profiler_module "github.com/x0k/ps2-spy/internal/profiler"
)

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" env:"METRICS_ENABLED"`
	Address string `yaml:"address" env:"METRICS_ADDRESS"`
}

type Config struct {
	Logger  LoggerConfig  `yaml:"logger"`
	Metrics MetricsConfig `yaml:"metrics"`

	Profiler profiler_module.Config `yaml:"profiler"`
}

func MustLoadConfig(configPath string) *Config {
	cfg := &Config{}
	var cfgErr error
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfgErr = cleanenv.ReadEnv(cfg)
	} else if err == nil {
		cfgErr = cleanenv.ReadConfig(configPath, cfg)
	} else {
		cfgErr = err
	}
	if cfgErr != nil {
		log.Fatalf("cannot read config: %s", cfgErr)
	}
	return cfg
}
