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
	core.MessageBase
	Connected core.StrBool `json:"connected"`
}

func IsConnectionStateChangedMessage(msg core.MessageBase) bool {
	return msg.Service == core.PushService && MessageType(msg.Type) == ConnectionStateChangedType
}

type ServiceStateChanged struct {
	core.MessageBase
	Detail string       `json:"detail"`
	Online core.StrBool `json:"online"`
}

func (s ServiceStateChanged) Type() MessageType {
	return ServiceStateChangedType
}

type Heartbeat struct {
	core.MessageBase
	Timestamp string                  `json:"timestamp"`
	Online    map[string]core.StrBool `json:"online"`
}

func (h Heartbeat) Type() MessageType {
	return HeartbeatType
}

type ServiceMessage[T any] struct {
	core.MessageBase
	Payload T `json:"payload"`
}

func (s ServiceMessage[T]) Type() MessageType {
	return ServiceMessageType
}

type SubscriptionSettings struct {
	Characters                     []string `json:"characters"`
	CharactersCount                int      `json:"charactersCount"`
	EventNames                     []string `json:"eventNames"`
	Worlds                         []string `json:"worlds"`
	LogicalAndCharactersWithWorlds bool     `json:"logicalAndCharactersWithWorlds"`
}

func (s SubscriptionSettings) Type() MessageType {
	return SubscriptionSignatureField
}

const SubscriptionSignatureField = "subscription"

type Subscription struct {
	Subscription SubscriptionSettings `json:"subscription"`
}

const HelpSignatureField = "send this for help"

type Help struct {
	Help core.CommandBase `json:"send this for help"`
}
