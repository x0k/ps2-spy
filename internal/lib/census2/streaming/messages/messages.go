package messages

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
	Detail string       `json:"detail"`
	Online core.StrBool `json:"online"`
}

type Heartbeat struct {
	core.MessageBase
	Timestamp string                  `json:"timestamp"`
	Online    map[string]core.StrBool `json:"online"`
}

type ServiceMessage[T any] struct {
	core.MessageBase
	Payload T `json:"payload"`
}

type SubscriptionBase struct {
	EventNames                     []string `json:"eventNames"`
	Worlds                         []string `json:"worlds"`
	LogicalAndCharactersWithWorlds bool     `json:"logicalAndCharactersWithWorlds"`
}

type AllCharactersSubscription struct {
	SubscriptionBase
	Characters []string `json:"characters"`
}

type CharactersCountSubscription struct {
	SubscriptionBase
	CharactersCount int `json:"charactersCount"`
}

type Subscription[S any] struct {
	Subscription S `json:"subscription"`
}

type Help struct {
	Help core.CommandBase `json:"send this for help"`
}
