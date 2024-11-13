package discord_module

import "time"

type Config struct {
	Token                 string        `yaml:"token" env:"DISCORD_TOKEN"`
	RemoveCommands        bool          `yaml:"remove_commands" env:"DISCORD_REMOVE_COMMANDS"`
	CommandHandlerTimeout time.Duration `yaml:"command_handler_timeout" env:"DISCORD_COMMAND_HANDLER_TIMEOUT"`
	EventHandlerTimeout   time.Duration `yaml:"event_handler_timeout" env:"DISCORD_EVENT_HANDLER_TIMEOUT"`
}
