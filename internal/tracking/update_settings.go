package tracking

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

var ErrTooManyOutfits = fmt.Errorf("too many outfits")
var ErrTooManyCharacters = fmt.Errorf("too many characters")

type ErrFailedToIdentifyEntities struct {
	OutfitTags     []string
	FoundOutfitIds map[string]ps2.OutfitId
	CharNames      []string
	FoundCharIds   map[string]ps2.CharacterId
}

func (e ErrFailedToIdentifyEntities) Error() string {
	return "failed to identify entities"
}

type Settings struct {
	Characters []ps2.CharacterId
	Outfits    []ps2.OutfitId
}

type SettingsRepo interface {
	Settings(context.Context, discord.ChannelId, ps2_platforms.Platform) (Settings, error)
	Delete(context.Context, discord.ChannelId, ps2_platforms.Platform, []ps2.OutfitId, []ps2.CharacterId) error
	Save(context.Context, discord.ChannelId, ps2_platforms.Platform, []ps2.OutfitId, []ps2.CharacterId) error
	Transaction(ctx context.Context, f func(r SettingsRepo) error) error
}

type OutfitsRepo interface {
	OutfitsByTag(context.Context, ps2_platforms.Platform, []string) (map[string]ps2.OutfitId, error)
}

type CharactersRepo interface {
	CharactersByName(context.Context, ps2_platforms.Platform, []string) (map[string]ps2.CharacterId, error)
}

type SettingsUpdater struct {
	repo                      SettingsRepo
	outfitsRepo               OutfitsRepo
	charactersRepo            CharactersRepo
	maxNumOfTrackedOutfits    int
	maxNumOfTrackedCharacters int
	publisher                 pubsub.Publisher[Event]
}

func NewSettingsUpdater(
	repo SettingsRepo,
	outfitsRepo OutfitsRepo,
	charactersRepo CharactersRepo,
	maxNumOfTrackedOutfits int,
	maxNumOfTrackedCharacters int,
	publisher pubsub.Publisher[Event],
) *SettingsUpdater {
	return &SettingsUpdater{
		repo:                      repo,
		outfitsRepo:               outfitsRepo,
		charactersRepo:            charactersRepo,
		maxNumOfTrackedOutfits:    maxNumOfTrackedOutfits,
		maxNumOfTrackedCharacters: maxNumOfTrackedCharacters,
		publisher:                 publisher,
	}
}

func (s *SettingsUpdater) Update(
	ctx context.Context,
	channelId discord.ChannelId,
	platform ps2_platforms.Platform,
	outfitTags []string,
	charNames []string,
) error {
	if len(outfitTags) > s.maxNumOfTrackedOutfits {
		return ErrTooManyOutfits
	}
	if len(charNames) > s.maxNumOfTrackedCharacters {
		return ErrTooManyCharacters
	}

	outfitIds, _ := s.outfitsRepo.OutfitsByTag(ctx, platform, outfitTags)
	charIds, _ := s.charactersRepo.CharactersByName(ctx, platform, charNames)

	if len(outfitTags) > len(outfitIds) || len(charNames) > len(charIds) {
		return ErrFailedToIdentifyEntities{
			OutfitTags:     outfitTags,
			FoundOutfitIds: outfitIds,
			CharNames:      charNames,
			FoundCharIds:   charIds,
		}
	}

	var outfitsDiff diff.Diff[ps2.OutfitId]
	var charactersDiff diff.Diff[ps2.CharacterId]
	err := s.repo.Transaction(ctx, func(r SettingsRepo) error {
		settings, err := r.Settings(channelId, platform)
		if err != nil {
			return fmt.Errorf("failed to get settings: %w", err)
		}
		outfitsDiff = diff.SliceAndMapValuesDiff(settings.Outfits, outfitIds)
		charactersDiff = diff.SliceAndMapValuesDiff(settings.Characters, charIds)
		if outfitsDiff.IsEmpty() && charactersDiff.IsEmpty() {
			return nil
		}
		if err := r.Delete(channelId, platform, outfitsDiff.ToDel, charactersDiff.ToDel); err != nil {
			return fmt.Errorf("failed to delete settings: %w", err)
		}
		if err := r.Save(channelId, platform, outfitsDiff.ToAdd, charactersDiff.ToAdd); err != nil {
			return fmt.Errorf("failed to save settings: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}
	s.publisher.Publish(TrackingSettingsUpdated{
		ChannelId:  channelId,
		Platform:   platform,
		Outfits:    outfitsDiff,
		Characters: charactersDiff,
	})
	return nil
}
