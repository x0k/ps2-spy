package discord_module

import (
	"time"
)

type Config struct {
	Token                 string
	RemoveCommands        bool
	CommandHandlerTimeout time.Duration
	EventHandlerTimeout   time.Duration
	Commands              []*Command
}
