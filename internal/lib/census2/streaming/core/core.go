package core

type StrBool string

const (
	True  StrBool = "true"
	False StrBool = "false"
)

const (
	EventService = "event"
)

type CommandBase struct {
	Service string `json:"service"`
	Action  string `json:"action"`
}

type MessageBase struct {
	Service string `json:"service"`
	Type    string `json:"type"`
}
