package streaming

import (
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

type MessageType string

type Message = pubsub.Event[MessageType]

const (
	ConnectionStateChangedType MessageType = "connectionStateChanged"
	ServiceStateChangedType    MessageType = "serviceStateChanged"
	HeartbeatType              MessageType = "heartbeat"
	ServiceMessageType         MessageType = "serviceMessage"
)

type ConnectionStateChanged struct {
	core.MessageBase `mapstructure:",squash"`
	Connected        core.StrBool `json:"connected" mapstructure:"connected"`
}

func IsConnectionStateChangedMessage(msg core.MessageBase) bool {
	return msg.Service == core.PushService && MessageType(msg.Type) == ConnectionStateChangedType
}

type ServiceStateChanged struct {
	core.MessageBase `mapstructure:",squash"`
	Detail           string       `json:"detail" mapstructure:"detail"`
	Online           core.StrBool `json:"online" mapstructure:"online"`
}

func (s *ServiceStateChanged) Type() MessageType {
	return ServiceStateChangedType
}

type Heartbeat struct {
	core.MessageBase `mapstructure:",squash"`
	Timestamp        string                  `json:"timestamp" mapstructure:"timestamp"`
	Online           map[string]core.StrBool `json:"online" mapstructure:"online"`
}

func (h *Heartbeat) Type() MessageType {
	return HeartbeatType
}

type ServiceMessage[T any] struct {
	core.MessageBase `mapstructure:",squash"`
	Payload          T `json:"payload" mapstructure:"payload"`
}

func (s *ServiceMessage[T]) Type() MessageType {
	return ServiceMessageType
}

type SubscriptionSettings struct {
	Characters                     []string `json:"characters" mapstructure:"characters"`
	CharactersCount                int      `json:"charactersCount" mapstructure:"charactersCount"`
	EventNames                     []string `json:"eventNames" mapstructure:"eventNames"`
	Worlds                         []string `json:"worlds" mapstructure:"worlds"`
	LogicalAndCharactersWithWorlds bool     `json:"logicalAndCharactersWithWorlds" mapstructure:"logicalAndCharactersWithWorlds"`
}

func (s *SubscriptionSettings) Type() MessageType {
	return SubscriptionSignatureField
}

const SubscriptionSignatureField = "subscription"

type Subscription struct {
	Subscription SubscriptionSettings `json:"subscription" mapstructure:"subscription"`
}

const HelpSignatureField = "send this for help"

type Help struct {
	Help core.CommandBase `json:"send this for help" mapstructure:"send this for help"`
}
