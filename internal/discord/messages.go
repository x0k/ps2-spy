package discord

import (
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/message"
)

type Error struct {
	Msg string
	Err error
}

type Message = func(*message.Printer) (string, *Error)

type Edit = func(*message.Printer) (*discordgo.WebhookEdit, *Error)

type Response = func(*message.Printer) (*discordgo.InteractionResponseData, *Error)

type FollowUp = func(*message.Printer) (*discordgo.WebhookParams, *Error)
