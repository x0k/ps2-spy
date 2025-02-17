package discord_event_handlers

import (
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func New(
	m *HandlersManager,
	messages *discord_messages.Messages,
	onlineTrackableEntitiesCountLoader OnlineTrackableEntitiesCountLoader,
	channelTitleUpdater ChannelTitleUpdater,
	trackingSettingsDiffViewLoader TrackingSettingsDiffViewLoader,
) []Handler {
	return []Handler{
		NewChannelLanguageSaved(m, messages, onlineTrackableEntitiesCountLoader, channelTitleUpdater),
		NewChannelTitleUpdatesSaved(m, messages, onlineTrackableEntitiesCountLoader, channelTitleUpdater),
		NewChannelTrackerStarted(m, messages),
		NewChannelTrackerStopped(m, messages),
		NewTrackingSettingsUpdateHandler(m, messages, trackingSettingsDiffViewLoader),
	}
}

func NewPlatform(
	m *HandlersManager,
	messages *discord_messages.Messages,
	platform ps2_platforms.Platform,
	outfitLoader OutfitLoader,
	facilityLoader FacilityLoader,
	charactersLoader CharactersLoader,
	characterLoader CharacterLoader,
	onlineTrackableEntitiesCountLoader OnlineTrackableEntitiesCountLoader,
	channelTitleUpdater ChannelTitleUpdater,
) []Handler {
	return []Handler{
		NewFacilityControl(m, messages, outfitLoader, facilityLoader, platform),
		NewFacilityLoss(m, messages, outfitLoader, facilityLoader, platform),
		NewOutfitMembersUpdate(m, messages, outfitLoader, charactersLoader, platform),
		NewPlayerLogin(m, messages, onlineTrackableEntitiesCountLoader, channelTitleUpdater),
		NewPlayerFakeLogin(m, messages, onlineTrackableEntitiesCountLoader, channelTitleUpdater),
		NewPlayerLogout(m, messages, characterLoader, onlineTrackableEntitiesCountLoader, channelTitleUpdater),
	}
}
