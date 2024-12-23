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
	Token                 string        `yaml:"token" env:"DISCORD_TOKEN" env-required:"true"`
	RemoveCommands        bool          `yaml:"remove_commands" env:"DISCORD_REMOVE_COMMANDS"`
	CommandHandlerTimeout time.Duration `yaml:"command_handler_timeout" env:"DISCORD_COMMAND_HANDLER_TIMEOUT" env-default:"25s"`
	EventHandlerTimeout   time.Duration `yaml:"event_handler_timeout" env:"DISCORD_EVENT_HANDLER_TIMEOUT" env-default:"5m"`
}

type StorageConfig struct {
	Path           string `yaml:"path" env:"STORAGE_PATH" env-required:"true"`
	MigrationsPath string `yaml:"migrations_path" env:"STORAGE_MIGRATIONS_PATH" env-required:"true"`
}

type HttpClientConfig struct {
	Timeout               time.Duration `yaml:"timeout" env:"HTTP_CLIENT_TIMEOUT" env-default:"30s"`
	KeepAlive             time.Duration `yaml:"keep_alive" env:"HTTP_CLIENT_KEEP_ALIVE" env-default:"30s"`
	TLSHandshakeTimeout   time.Duration `yaml:"tls_handshake_timeout" env:"HTTP_CLIENT_TLS_HANDSHAKE_TIMEOUT" env-default:"10s"`
	ResponseHeaderTimeout time.Duration `yaml:"response_header_timeout" env:"HTTP_CLIENT_RESPONSE_HEADER_TIMEOUT" env-default:"10s"`
	ExpectContinueTimeout time.Duration `yaml:"expect_continue_timeout" env:"HTTP_CLIENT_EXPECT_CONTINUE_TIMEOUT" env-default:"1s"`
}

type CensusConfig struct {
	ServiceId         string `yaml:"service_id" env:"CENSUS_SERVICE_ID" env-required:"true"`
	StreamingEndpoint string `yaml:"streaming_endpoint" env:"CENSUS_STREAMING_ENDPOINT" env-default:"wss://push.planetside2.com/streaming"`
}

type StatsTrackerConfig struct {
	MaxTrackingDuration time.Duration `yaml:"max_tracking_duration" env:"STATS_TRACKER_MAX_TRACKING_DURATION" env-default:"4h"`
}

type Config struct {
	AppName string `yaml:"app_name" env:"APP_NAME" env-default:"PS 2 Spy"`

	Logger       LoggerConfig       `yaml:"logger"`
	Profiler     ProfilerConfig     `yaml:"profiler"`
	Discord      DiscordConfig      `yaml:"discord"`
	Metrics      MetricsConfig      `yaml:"metrics"`
	Storage      StorageConfig      `yaml:"storage"`
	HttpClient   HttpClientConfig   `yaml:"http_client"`
	Census       CensusConfig       `yaml:"census"`
	StatsTracker StatsTrackerConfig `yaml:"stats_tracker"`
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
