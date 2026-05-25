package settings

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/lib/mapx"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/tracking"
)

type Opts struct {
	SettingsRepo         SettingsRepo
	OutfitsRepo          OutfitsRepo
	CharactersRepo       CharactersRepo
	TrackingRepo         TrackingRepo
	MaxTrackedOutfits    int
	MaxTrackedCharacters int
	Publisher            pubsub.Publisher[tracking.Event]
	StoragePublisher     pubsub.Publisher[storage.Event]
}

type Service struct {
	settingsRepo         SettingsRepo
	outfitsRepo          OutfitsRepo
	charactersRepo       CharactersRepo
	trackingRepo         TrackingRepo
	maxTrackedOutfits    int
	maxTrackedCharacters int
	publisher            pubsub.Publisher[tracking.Event]
	storagePublisher     pubsub.Publisher[storage.Event]
}

func New(opts Opts) *Service {
	return &Service{
		settingsRepo:         opts.SettingsRepo,
		outfitsRepo:          opts.OutfitsRepo,
		charactersRepo:       opts.CharactersRepo,
		trackingRepo:         opts.TrackingRepo,
		maxTrackedOutfits:    opts.MaxTrackedOutfits,
		maxTrackedCharacters: opts.MaxTrackedCharacters,
		publisher:            opts.Publisher,
		storagePublisher:     opts.StoragePublisher,
	}
}

func (s *Service) Load(
	ctx context.Context, channelId discord.ChannelId, platform ps2_platforms.Platform,
) (tracking.SettingsData, error) {
	settings, err := s.settingsRepo.Get(ctx, channelId, platform)
	if err != nil {
		return tracking.SettingsData{}, fmt.Errorf("failed to load settings: %w", err)
	}
	members, err := s.trackingRepo.OnlineOutfitMembers(ctx, platform, settings.Outfits)
	if err != nil {
		return tracking.SettingsData{}, fmt.Errorf("failed to load outfit members: %w", err)
	}
	characters, err := s.trackingRepo.OnlineCharacters(ctx, platform, settings.Characters)
	if err != nil {
		return tracking.SettingsData{}, fmt.Errorf("failed to load characters: %w", err)
	}
	outfits := make(map[ps2.OutfitId][]ps2.Character, len(members))
	for outfitId, m := range members {
		outfits[outfitId] = mapx.Values(m)
	}
	return tracking.SettingsData{
		Characters: mapx.Values(characters),
		Outfits:    outfits,
	}, nil
}

func (s *Service) LoadView(ctx context.Context, channelId discord.ChannelId, platform ps2_platforms.Platform) (tracking.SettingsView, error) {
	settings, err := s.settingsRepo.Get(ctx, channelId, platform)
	if err != nil {
		return tracking.SettingsView{}, fmt.Errorf("failed to load settings: %w", err)
	}
	outfits, err := s.outfitsRepo.OutfitTagsByIds(ctx, platform, settings.Outfits)
	if err != nil {
		return tracking.SettingsView{}, fmt.Errorf("failed to load outfits: %w", err)
	}
	characters, err := s.charactersRepo.CharacterNamesByIds(ctx, platform, settings.Characters)
	if err != nil {
		return tracking.SettingsView{}, fmt.Errorf("failed to load characters: %w", err)
	}
	return tracking.SettingsView{
		Outfits:    mapx.Values(outfits),
		Characters: mapx.Values(characters),
	}, nil
}

func (s *Service) LoadDiffView(
	ctx context.Context, platform ps2_platforms.Platform, d tracking.SettingsDiff,
) (tracking.SettingsDiffView, error) {
	charIds := make([]ps2.CharacterId, 0, len(d.Characters.ToAdd)+len(d.Characters.ToDel))
	charIds = append(charIds, d.Characters.ToAdd...)
	charIds = append(charIds, d.Characters.ToDel...)
	outfitIds := make([]ps2.OutfitId, 0, len(d.Outfits.ToAdd)+len(d.Outfits.ToDel))
	outfitIds = append(outfitIds, d.Outfits.ToAdd...)
	outfitIds = append(outfitIds, d.Outfits.ToDel...)

	charNames, err := s.charactersRepo.CharacterNamesByIds(ctx, platform, charIds)
	if err != nil {
		return tracking.SettingsDiffView{}, fmt.Errorf("failed to load character names: %w", err)
	}
	outfitTags, err := s.outfitsRepo.OutfitTagsByIds(ctx, platform, outfitIds)
	if err != nil {
		return tracking.SettingsDiffView{}, fmt.Errorf("failed to load outfit tags: %w", err)
	}
	diffView := tracking.SettingsDiffView{
		Characters: diff.Diff[string]{
			ToAdd: make([]string, 0, len(d.Characters.ToAdd)),
			ToDel: make([]string, 0, len(d.Characters.ToDel)),
		},
		Outfits: diff.Diff[string]{
			ToAdd: make([]string, 0, len(d.Outfits.ToAdd)),
			ToDel: make([]string, 0, len(d.Outfits.ToDel)),
		},
	}
	for _, charId := range d.Characters.ToAdd {
		name, ok := charNames[charId]
		if !ok {
			return tracking.SettingsDiffView{}, fmt.Errorf("character %s not found", charId)
		}
		diffView.Characters.ToAdd = append(diffView.Characters.ToAdd, name)
	}
	for _, charId := range d.Characters.ToDel {
		name, ok := charNames[charId]
		if !ok {
			return tracking.SettingsDiffView{}, fmt.Errorf("character %s not found", charId)
		}
		diffView.Characters.ToDel = append(diffView.Characters.ToDel, name)
	}
	for _, outfitId := range d.Outfits.ToAdd {
		tag, ok := outfitTags[outfitId]
		if !ok {
			return tracking.SettingsDiffView{}, fmt.Errorf("outfit %s not found", outfitId)
		}
		diffView.Outfits.ToAdd = append(diffView.Outfits.ToAdd, tag)
	}
	for _, outfitId := range d.Outfits.ToDel {
		tag, ok := outfitTags[outfitId]
		if !ok {
			return tracking.SettingsDiffView{}, fmt.Errorf("outfit %s not found", outfitId)
		}
		diffView.Outfits.ToDel = append(diffView.Outfits.ToDel, tag)
	}
	return diffView, nil
}

func (s *Service) Update(
	ctx context.Context,
	channelId discord.ChannelId,
	platform ps2_platforms.Platform,
	settings tracking.SettingsView,
	updater discord.UserId,
) error {
	if len(settings.Outfits) > s.maxTrackedOutfits {
		return tracking.ErrTooManyOutfits(settings)
	}
	if len(settings.Characters) > s.maxTrackedCharacters {
		return tracking.ErrTooManyCharacters(settings)
	}

	outfitIds, _ := s.outfitsRepo.OutfitIdsByTags(ctx, platform, settings.Outfits)
	charIds, _ := s.charactersRepo.CharacterIdsByNames(ctx, platform, settings.Characters)

	if len(settings.Outfits) > len(outfitIds) || len(settings.Characters) > len(charIds) {
		return tracking.ErrFailedToIdentifyEntities{
			OutfitTags:     settings.Outfits,
			FoundOutfitIds: outfitIds,
			CharNames:      settings.Characters,
			FoundCharIds:   charIds,
		}
	}

	settingsDiff, err := s.settingsRepo.Update(ctx, channelId, platform, tracking.Settings{
		Characters: mapx.Values(charIds),
		Outfits:    mapx.Values(outfitIds),
	})
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	if !settingsDiff.IsEmpty() {
		s.storagePublisher.Publish(storage.ChannelTrackingSettingsSaved{
			ChannelId: channelId,
			Platform:  platform,
			Updater:   updater,
		})
		s.publisher.Publish(tracking.TrackingSettingsUpdated{
			ChannelId: channelId,
			Platform:  platform,
			Diff:      settingsDiff,
			Updater:   updater,
		})
	}

	return nil
}
