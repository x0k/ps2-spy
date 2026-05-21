package tracking

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func testLogger() *logger.Logger {
	return logger.New(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestPlatformManagerTracksDuplicatedCharactersAndOutfitMembers(t *testing.T) {
	ctx := context.Background()
	const (
		characterId       = ps2.CharacterId("char-1")
		outfitMemberId    = ps2.CharacterId("member-1")
		outfitId          = ps2.OutfitId("outfit-1")
		trackingChannelId = discord.ChannelId("tracking-channel")
	)
	pm := newPlatformManager(
		testLogger(),
		ps2_platforms.PC,
		func(ctx context.Context, platform ps2_platforms.Platform, id ps2.CharacterId) (ps2.Character, error) {
			return ps2.Character{Id: id, Platform: platform}, nil
		},
		func(ctx context.Context, platform ps2_platforms.Platform, character ps2.Character) ([]discord.Channel, error) {
			return []discord.Channel{{Id: trackingChannelId}}, nil
		},
		func(context.Context, ps2_platforms.Platform) ([]ps2.CharacterId, error) {
			return []ps2.CharacterId{characterId, characterId}, nil
		},
		func(context.Context, ps2_platforms.Platform, ps2.OutfitId) ([]ps2.CharacterId, error) {
			return nil, nil
		},
		func(ctx context.Context, platform ps2_platforms.Platform, id ps2.OutfitId) ([]discord.Channel, error) {
			return []discord.Channel{{Id: trackingChannelId}}, nil
		},
		func(context.Context, ps2_platforms.Platform) ([]ps2.OutfitId, error) {
			return []ps2.OutfitId{outfitId, outfitId}, nil
		},
	)

	pm.rebuildFilters(ctx)
	pm.wg.Wait()

	channels, err := pm.ChannelsForCharacter(ctx, characterId)
	if err != nil {
		t.Fatalf("ChannelsForCharacter() error = %v", err)
	}
	if len(channels) != 1 || channels[0].Id != trackingChannelId {
		t.Fatalf("ChannelsForCharacter() = %#v", channels)
	}

	pm.TrackOutfitMembers(outfitId, []ps2.CharacterId{outfitMemberId})
	channels, err = pm.ChannelsForCharacter(ctx, outfitMemberId)
	if err != nil {
		t.Fatalf("ChannelsForCharacter(outfit member) error = %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected outfit member to be tracked, got %#v", channels)
	}

	pm.UntrackOutfitMembers(outfitId, []ps2.CharacterId{outfitMemberId})
	channels, err = pm.ChannelsForCharacter(ctx, outfitMemberId)
	if err != nil {
		t.Fatalf("ChannelsForCharacter(untracked outfit member) error = %v", err)
	}
	if len(channels) != 0 {
		t.Fatalf("expected outfit member to be untracked, got %#v", channels)
	}
}

func TestPlatformManagerHandlesTrackingSettingsUpdate(t *testing.T) {
	ctx := context.Background()
	const (
		characterId    = ps2.CharacterId("char-1")
		outfitMemberId = ps2.CharacterId("member-1")
		outfitId       = ps2.OutfitId("outfit-1")
	)
	pm := newPlatformManager(
		testLogger(),
		ps2_platforms.PC,
		func(ctx context.Context, platform ps2_platforms.Platform, id ps2.CharacterId) (ps2.Character, error) {
			return ps2.Character{Id: id, Platform: platform}, nil
		},
		func(ctx context.Context, platform ps2_platforms.Platform, character ps2.Character) ([]discord.Channel, error) {
			return []discord.Channel{{Id: "channel"}}, nil
		},
		func(context.Context, ps2_platforms.Platform) ([]ps2.CharacterId, error) {
			return nil, nil
		},
		func(context.Context, ps2_platforms.Platform, ps2.OutfitId) ([]ps2.CharacterId, error) {
			return []ps2.CharacterId{outfitMemberId}, nil
		},
		func(ctx context.Context, platform ps2_platforms.Platform, id ps2.OutfitId) ([]discord.Channel, error) {
			return []discord.Channel{{Id: "channel"}}, nil
		},
		func(context.Context, ps2_platforms.Platform) ([]ps2.OutfitId, error) {
			return nil, nil
		},
	)

	pm.HandleTrackingSettingsUpdate(ctx, TrackingSettingsUpdated{
		Diff: SettingsDiff{
			Characters: diff.Diff[ps2.CharacterId]{ToAdd: []ps2.CharacterId{characterId}},
			Outfits:    diff.Diff[ps2.OutfitId]{ToAdd: []ps2.OutfitId{outfitId}},
		},
	})
	pm.wg.Wait()

	for _, id := range []ps2.CharacterId{characterId, outfitMemberId} {
		channels, err := pm.ChannelsForCharacter(ctx, id)
		if err != nil {
			t.Fatalf("ChannelsForCharacter(%s) error = %v", id, err)
		}
		if len(channels) != 1 {
			t.Fatalf("expected %s to be tracked, got %#v", id, channels)
		}
	}

	channels, err := pm.ChannelsForOutfit(ctx, outfitId)
	if err != nil {
		t.Fatalf("ChannelsForOutfit() error = %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected outfit to be tracked, got %#v", channels)
	}
}
