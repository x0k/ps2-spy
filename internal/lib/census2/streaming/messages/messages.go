package ps2messages

import "github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"

const (
	ConnectionStateChangedType = "connectionStateChanged"
	ServiceStateChangedType    = "serviceStateChanged"
	HeartbeatType              = "heartbeat"
	ServiceMessageType         = "serviceMessage"
)

type ConnectionStateChanged struct {
	core.MessageBase
	Connected core.StrBool `json:"connected" mapstructure:"connected"`
}

func IsConnectionStateChangedMessage(msg core.MessageBase) bool {
	return msg.Service == core.PushService && msg.Type == ConnectionStateChangedType
}

type ServiceStateChanged struct {
	core.MessageBase
	Detail string       `json:"detail" mapstructure:"detail"`
	Online core.StrBool `json:"online" mapstructure:"online"`
}

type Heartbeat struct {
	core.MessageBase
	Timestamp string                  `json:"timestamp" mapstructure:"timestamp"`
	Online    map[string]core.StrBool `json:"online" mapstructure:"online"`
}

type ServiceMessage[T any] struct {
	core.MessageBase
	Payload T `json:"payload" mapstructure:"payload"`
}

type SubscriptionSettings struct {
	Characters                     []string `json:"characters" mapstructure:"characters"`
	CharactersCount                int      `json:"charactersCount" mapstructure:"charactersCount"`
	EventNames                     []string `json:"eventNames" mapstructure:"eventNames"`
	Worlds                         []string `json:"worlds" mapstructure:"worlds"`
	LogicalAndCharactersWithWorlds bool     `json:"logicalAndCharactersWithWorlds" mapstructure:"logicalAndCharactersWithWorlds"`
}

const SubscriptionSignatureField = "subscription"

type Subscription struct {
	Subscription SubscriptionSettings `json:"subscription" mapstructure:"subscription"`
}

const HelpSignatureField = "send this for help"

type Help struct {
	Help core.CommandBase `json:"send this for help" mapstructure:"send this for help"`
}
