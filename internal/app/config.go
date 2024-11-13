package app

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" env:"METRICS_ENABLED"`
	Address string `yaml:"address" env:"METRICS_ADDRESS"`
}

type ProfilerConfig struct {
	Enabled bool   `yaml:"enabled" env:"PROFILER_ENABLED"`
	Address string `yaml:"address" env:"PROFILER_ADDRESS"`
}

type DiscordConfig struct {
	Token                 string        `yaml:"token" env:"DISCORD_TOKEN"`
	RemoveCommands        bool          `yaml:"remove_commands" env:"DISCORD_REMOVE_COMMANDS"`
	CommandHandlerTimeout time.Duration `yaml:"command_handler_timeout" env:"DISCORD_COMMAND_HANDLER_TIMEOUT"`
	EventHandlerTimeout   time.Duration `yaml:"event_handler_timeout" env:"DISCORD_EVENT_HANDLER_TIMEOUT"`
}

type Config struct {
	Logger   LoggerConfig   `yaml:"logger"`
	Profiler ProfilerConfig `yaml:"profiler"`
	Discord  DiscordConfig  `yaml:"discord"`
	Metrics  MetricsConfig  `yaml:"metrics"`
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
