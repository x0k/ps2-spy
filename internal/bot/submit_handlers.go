package bot

import (
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	channelsetup "github.com/x0k/ps2-spy/internal/bot/handlers/submit/channel_setup"
	"github.com/x0k/ps2-spy/internal/loaders"
)

func NewSubmitHandlers(
	characterIdsLoader loaders.QueriedLoader[[]string, []string],
	pcSaver channelsetup.Saver,
	ps4euSave channelsetup.Saver,
	ps4usSave channelsetup.Saver,
) map[string]handlers.InteractionHandler {
	return map[string]handlers.InteractionHandler{
		handlers.CHANNEL_SETUP_PC_MODAL:     channelsetup.New(characterIdsLoader, pcSaver),
		handlers.CHANNEL_SETUP_PS4_EU_MODAL: channelsetup.New(characterIdsLoader, ps4euSave),
		handlers.CHANNEL_SETUP_PS4_US_MODAL: channelsetup.New(characterIdsLoader, ps4usSave),
	}
}
