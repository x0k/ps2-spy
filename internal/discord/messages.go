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

type Chunkable struct {
	Chunks int
	Print  func(start int, count int) (string, *Error)
}

func NewChunkableMessage(
	chunks int,
	print func(p *message.Printer, start int, count int) (string, *Error),
) ChunkableMessage {
	return func(p *message.Printer) Chunkable {
		return Chunkable{
			Chunks: chunks,
			Print: func(start, count int) (string, *Error) {
				return print(p, start, count)
			},
		}
	}
}

type ChunkableMessage = func(*message.Printer) Chunkable

type Edit = func(*message.Printer) (*discordgo.WebhookEdit, *Error)

type Response = func(*message.Printer) (*discordgo.InteractionResponseData, *Error)

type FollowUp = func(*message.Printer) (*discordgo.WebhookParams, *Error)
