package tracking

import (
	"context"
	"fmt"

	uow "github.com/x0k/ps2-spy/internal/adapters/unit_of_work"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

var ErrTooManyOutfits = fmt.Errorf("too many outfits")
var ErrTooManyCharacters = fmt.Errorf("too many characters")

type Settings struct {
	Characters []ps2.CharacterId
	Outfits    []ps2.OutfitId
}

type SettingsRepo[T any] interface {
	Settings() (Settings, error)
	Delete(uow.UnitOfWork[T, Event], discord.ChannelId, ps2_platforms.Platform, []ps2.OutfitId, []ps2.CharacterId) error
	Save(uow.UnitOfWork[T, Event], discord.ChannelId, ps2_platforms.Platform, []ps2.OutfitId, []ps2.CharacterId) error
}

type OutfitsRepo interface {
	OutfitsByTag(context.Context, ps2_platforms.Platform, []string) ([]ps2.OutfitId, error)
}

type CharactersRepo interface {
	CharactersByName(context.Context, ps2_platforms.Platform, []string) ([]ps2.CharacterId, error)
}

type SettingsUpdater[T any] struct {
	repo                      SettingsRepo[T]
	outfitsRepo               OutfitsRepo
	charactersRepo            CharactersRepo
	newUow                    uow.Factory[T, Event]
	maxNumOfTrackedOutfits    int
	maxNumOfTrackedCharacters int
}

func (s *SettingsUpdater[T]) Update(
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
	outfitIds, err := s.outfitsRepo.OutfitsByTag(ctx, platform, outfitTags)
	if err != nil {
		return fmt.Errorf("failed to get outfits: %w", err)
	}
	charIds, err := s.charactersRepo.CharactersByName(ctx, platform, charNames)
	if err != nil {
		return fmt.Errorf("failed to get characters: %w", err)
	}
	settings, err := s.repo.Settings()
	if err != nil {
		return fmt.Errorf("failed to get settings: %w", err)
	}
	outfitsDiff := diff.SlicesDiff(settings.Outfits, outfitIds)
	charactersDiff := diff.SlicesDiff(settings.Characters, charIds)
	if outfitsDiff.IsEmpty() && charactersDiff.IsEmpty() {
		return nil
	}
	tx, err := s.newUow(ctx)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	if err := s.repo.Delete(tx, channelId, platform, outfitsDiff.ToDel, charactersDiff.ToDel); err != nil {
		return fmt.Errorf("failed to delete settings: %w", err)
	}
	if err := s.repo.Save(tx, channelId, platform, outfitsDiff.ToAdd, charactersDiff.ToAdd); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}
	tx.Publish(TrackingSettingsUpdated{
		ChannelId:  channelId,
		Platform:   platform,
		Outfits:    outfitsDiff,
		Characters: charactersDiff,
	})
	return tx.Commit(ctx)
}
