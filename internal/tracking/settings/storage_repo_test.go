package settings

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/migrator"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage"
	sql_storage "github.com/x0k/ps2-spy/internal/storage/sql"
	"github.com/x0k/ps2-spy/internal/tracking"

	_ "modernc.org/sqlite"
)

func testLogger() *logger.Logger {
	return logger.New(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func newTestStorage(t *testing.T, ctx context.Context) storage.Storage {
	t.Helper()
	migrationsDir, err := filepath.Abs(filepath.Join("..", "..", "..", "db", "migrations"))
	if err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(t.TempDir(), "test.db")

	m := migrator.New(slog.New(slog.NewTextHandler(io.Discard, nil)), "sqlite://"+dbPath, "file://"+migrationsDir)
	if err := m.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	store := sql_storage.New(testLogger(), "sqlite://"+dbPath, pubsub.New[storage.EventType]())
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(ctx) })
	return store
}

func TestRepositoryUpdateCreatesChannelRow(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(newTestStorage(t, ctx), testLogger())

	diff, err := repo.Update(ctx, "test-channel", ps2_platforms.PC, tracking.Settings{
		Characters: []ps2.CharacterId{"char-1"},
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if len(diff.Characters.ToAdd) != 1 {
		t.Fatal("expected 1 char added")
	}

	channels, err := repo.TrackingChannelsForCharacter(ctx, ps2_platforms.PC, "char-1", "")
	if err != nil {
		t.Fatalf("TrackingChannelsForCharacter() error = %v", err)
	}
	if len(channels) != 1 {
		t.Fatal("TrackingChannelsForCharacter did not find the channel — channel row missing")
	}
	if channels[0].Id != discord.ChannelId("test-channel") {
		t.Fatalf("expected channel 'test-channel', got %q", channels[0].Id)
	}
}

func TestRepositoryTrackingChannelsForCharacterAfterUpdate(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(newTestStorage(t, ctx), testLogger())

	_, err := repo.Update(ctx, "test-channel", ps2_platforms.PC, tracking.Settings{
		Characters: []ps2.CharacterId{"char-alpha"},
		Outfits:    []ps2.OutfitId{"outfit-beta"},
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	channels, err := repo.TrackingChannelsForCharacter(ctx, ps2_platforms.PC, "char-alpha", "outfit-beta")
	if err != nil {
		t.Fatalf("TrackingChannelsForCharacter() error = %v", err)
	}
	if len(channels) != 1 || channels[0].Id != discord.ChannelId("test-channel") {
		t.Fatalf("expected [test-channel], got %+v", channels)
	}

	outfitChannels, err := repo.TrackingChannelsForOutfit(ctx, ps2_platforms.PC, "outfit-beta")
	if err != nil {
		t.Fatalf("TrackingChannelsForOutfit() error = %v", err)
	}
	if len(outfitChannels) != 1 {
		t.Fatalf("expected 1 channel for outfit, got %d", len(outfitChannels))
	}
}
