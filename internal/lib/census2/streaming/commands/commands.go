package ps2commands

import "github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"

const (
	EchoAction                    = "echo"
	SubscribeAction               = "subscribe"
	ClearSubscribeAction          = "clearSubscribe"
	RecentCharacterIdsAction      = "recentCharacterIds"
	RecentCharacterIdsCountAction = "recentCharacterIdsCount"
)

type EchoCommand[T any] struct {
	core.CommandBase
	Payload T `json:"payload"`
}

func Echo[T any](payload T) EchoCommand[T] {
	return EchoCommand[T]{
		CommandBase: core.CommandBase{
			Service: core.EventService,
			Action:  EchoAction,
		},
		Payload: payload,
	}
}

type SubscriptionSettings struct {
	Characters                     []string `json:"characters,omitempty"`
	EventNames                     []string `json:"eventNames,omitempty"`
	Worlds                         []string `json:"worlds,omitempty"`
	LogicalAndCharactersWithWorlds bool     `json:"logicalAndCharactersWithWorlds,omitempty"`
}

type ChangeSubscriptionCommand struct {
	core.CommandBase
	SubscriptionSettings
}

func Subscribe(payload SubscriptionSettings) ChangeSubscriptionCommand {
	return ChangeSubscriptionCommand{
		CommandBase: core.CommandBase{
			Service: core.EventService,
			Action:  SubscribeAction,
		},
		SubscriptionSettings: payload,
	}
}

func ClearSubscribe(payload SubscriptionSettings) ChangeSubscriptionCommand {
	return ChangeSubscriptionCommand{
		CommandBase: core.CommandBase{
			Service: core.EventService,
			Action:  ClearSubscribeAction,
		},
		SubscriptionSettings: payload,
	}
}

type ClearAllSubscribeCommand struct {
	core.CommandBase
	All core.StrBool `json:"all"`
}

func ClearAllSubscribe(all core.StrBool) ClearAllSubscribeCommand {
	return ClearAllSubscribeCommand{
		CommandBase: core.CommandBase{
			Service: core.EventService,
			Action:  ClearSubscribeAction,
		},
		All: all,
	}
}

func RecentCharacters() core.CommandBase {
	return core.CommandBase{
		Service: core.EventService,
		Action:  RecentCharacterIdsAction,
	}
}

func RecentCharactersCount() core.CommandBase {
	return core.CommandBase{
		Service: core.EventService,
		Action:  RecentCharacterIdsCountAction,
	}
}
