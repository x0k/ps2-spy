package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	TextHandler   = "text"
	JSONHandler   = "json"
	PrettyHandler = "pretty"
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

type StorageConfig struct {
	Path string `yaml:"path" env:"STORAGE_PATH" env-required:"true"`
}

type ProfilerConfig struct {
	Enabled bool   `yaml:"enabled" env:"PROFILER_ENABLED"`
	Address string `yaml:"address" env:"PROFILER_ADDRESS"`
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" env:"METRICS_ENABLED"`
	Address string `yaml:"address" env:"METRICS_ADDRESS"`
}

type Config struct {
	Logger   LoggerConfig   `yaml:"logger"`
	Storage  StorageConfig  `yaml:"storage"`
	Profiler ProfilerConfig `yaml:"profiler"`
	Metrics  MetricsConfig  `yaml:"metrics"`

	BotName                 string        `yaml:"bot_name" env:"BOT_NAME" env-default:"PS 2 Spy"`
	CommandHandlerTimeout   time.Duration `yaml:"command_handler_timeout" env:"COMMAND_HANDLER_TIMEOUT" env-default:"25s"`
	Ps2EventHandlerTimeout  time.Duration `yaml:"ps2_event_handler_timeout" env:"PS2_EVENT_HANDLER_TIMEOUT" env-default:"5m"`
	DiscordToken            string        `yaml:"token" env:"DISCORD_TOKEN" env-required:"true"`
	HttpClientTimeout       time.Duration `yaml:"http_client_timeout" env:"HTTP_CLIENT_TIMEOUT" env-default:"12s"`
	CensusServiceId         string        `yaml:"census_service_id" env:"CENSUS_SERVICE_ID" env-required:"true"`
	CensusStreamingEndpoint string        `yaml:"census_streaming_endpoint" env:"CENSUS_STREAMING_ENDPOINT" env-default:"wss://push.planetside2.com/streaming"`
	RemoveCommands          bool          `yaml:"remove_commands" env:"REMOVE_COMMANDS"`
}

func MustLoad(configPath string) *Config {
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
