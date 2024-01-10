package core

import "fmt"

type StrBool string

const (
	True  StrBool = "true"
	False StrBool = "false"
)

const (
	EventService = "event"
	PushService  = "push"
)

var ErrNoRequiredField = fmt.Errorf("no required field")
var ErrUnexpectedType = fmt.Errorf("unexpected type")

type CommandBase struct {
	Service string `json:"service"`
	Action  string `json:"action"`
}

const (
	ServiceMessageField = "service"
	TypeMessageField    = "type"
)

type MessageBase struct {
	Service string `json:"service" mapstructure:"service"`
	Type    string `json:"type" mapstructure:"type"`
}

func AsMessageBase(m map[string]any, b *MessageBase) error {
	serviceAny, ok := m[ServiceMessageField]
	if !ok {
		return fmt.Errorf("MessageBase.Service: %w", ErrNoRequiredField)
	}
	serviceStr, ok := serviceAny.(string)
	if !ok {
		return fmt.Errorf("MessageBase.Service: %w", ErrUnexpectedType)
	}
	b.Service = serviceStr
	typeAny, ok := m[TypeMessageField]
	if !ok {
		return fmt.Errorf("MessageBase.Type: %w", ErrNoRequiredField)
	}
	typeStr, ok := typeAny.(string)
	if !ok {
		return fmt.Errorf("MessageBase.Type: %w", ErrUnexpectedType)
	}
	b.Type = typeStr
	return nil
}

type EventBase struct {
	EventName string `json:"event_name"`
	Timestamp string `json:"timestamp"`
}
