package en_messages

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func (m *messages) SubscriptionSettingsOutfitTagsLoadError(outfitIds []ps2.OutfitId, platform ps2_platforms.Platform, err error) (*discordgo.WebhookEdit, *discord.Error) {
	return nil, &discord.Error{
		Msg: "Failed to load outfit tags (" + string(platform) + ")",
		Err: err,
	}
}
