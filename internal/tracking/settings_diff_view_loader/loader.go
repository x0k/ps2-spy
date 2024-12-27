package tracking_settings_diff_view_loader

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/tracking"
)

type OutfitsRepo interface {
	OutfitTagsByIds(context.Context, ps2_platforms.Platform, []ps2.OutfitId) (map[ps2.OutfitId]string, error)
}

type CharactersRepo interface {
	CharacterNamesByIds(context.Context, ps2_platforms.Platform, []ps2.CharacterId) (map[ps2.CharacterId]string, error)
}

type DiffViewLoader struct {
	outfitsRepo    OutfitsRepo
	charactersRepo CharactersRepo
}

func New(outfitsRepo OutfitsRepo, charactersRepo CharactersRepo) *DiffViewLoader {
	return &DiffViewLoader{
		outfitsRepo:    outfitsRepo,
		charactersRepo: charactersRepo,
	}
}

func (l *DiffViewLoader) Load(
	ctx context.Context, channelId discord.ChannelId, platform ps2_platforms.Platform, d tracking.SettingsDiff,
) (tracking.SettingsDiffView, error) {
	charIds := make([]ps2.CharacterId, 0, len(d.Characters.ToAdd)+len(d.Characters.ToDel))
	charIds = append(charIds, d.Characters.ToAdd...)
	charIds = append(charIds, d.Characters.ToDel...)
	outfitIds := make([]ps2.OutfitId, 0, len(d.Outfits.ToAdd)+len(d.Outfits.ToDel))
	outfitIds = append(outfitIds, d.Outfits.ToAdd...)
	outfitIds = append(outfitIds, d.Outfits.ToDel...)

	charNames, err := l.charactersRepo.CharacterNamesByIds(ctx, platform, charIds)
	if err != nil {
		return tracking.SettingsDiffView{}, fmt.Errorf("failed to load character names: %w", err)
	}
	outfitTags, err := l.outfitsRepo.OutfitTagsByIds(ctx, platform, outfitIds)
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
