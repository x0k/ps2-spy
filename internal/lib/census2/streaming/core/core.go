package core

import (
	"fmt"
)

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
	Service string `json:"service"`
	Type    string `json:"type"`
}

const EventNameField = "event_name"

type EventBase struct {
	EventName string `json:"event_name"`
	Timestamp string `json:"timestamp"`
}
