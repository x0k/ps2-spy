package sql_storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/db"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/shared"
	"github.com/x0k/ps2-spy/internal/storage"
	"golang.org/x/text/language"

	_ "modernc.org/sqlite"
)

type Storage struct {
	log                 *logger.Logger
	storagePath         string
	db                  *sql.DB
	queries             *db.Queries
	publisher           pubsub.Publisher[storage.Event]
	maxTrackingDuration time.Duration
}

func New(
	log *logger.Logger,
	storagePath string,
	maxTrackingDuration time.Duration,
	publisher pubsub.Publisher[storage.Event],
) *Storage {
	return &Storage{
		log:                 log,
		storagePath:         storagePath,
		maxTrackingDuration: maxTrackingDuration,
		publisher:           publisher,
		db:                  nil,
		queries:             nil,
	}
}

func (s *Storage) Open(ctx context.Context) error {
	var err error
	u, err := url.Parse(s.storagePath)
	if err != nil {
		return err
	}
	s.db, err = sql.Open(u.Scheme, u.Host+u.Path)
	if err != nil {
		return err
	}
	// s.db.SetMaxOpenConns(1)
	if _, err := s.db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return err
	}
	s.queries, err = db.Prepare(ctx, s.db)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) Close(_ context.Context) error {
	return errors.Join(
		s.queries.Close(),
		s.db.Close(),
	)
}

func (s *Storage) Queries() *db.Queries {
	return s.queries
}

func (s *Storage) Transaction(ctx context.Context, run func(s storage.Storage) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			s.log.Error(ctx, "failed to rollback transaction", sl.Err(err))
		}
	}()
	if err := run(&Storage{
		queries: s.queries.WithTx(tx),
	}); err != nil {
		return fmt.Errorf("failed to run transaction: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (s *Storage) Begin(
	ctx context.Context,
	expectedEventsCount int,
	run func(s *Storage) error,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			s.log.Error(ctx, "failed to rollback transaction", sl.Err(err))
		}
	}()
	bufferedPublisher := pubsub.NewBufferedPublisher(s.publisher, expectedEventsCount)
	tmp := &Storage{
		log:                 s.log,
		db:                  s.db,
		queries:             s.queries.WithTx(tx),
		publisher:           bufferedPublisher,
		storagePath:         s.storagePath,
		maxTrackingDuration: s.maxTrackingDuration,
	}
	err = run(tmp)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	bufferedPublisher.Flush()
	return nil
}

func (s *Storage) OutfitSynchronizedAt(ctx context.Context, platform ps2_platforms.Platform, outfitId ps2.OutfitId) (time.Time, error) {
	time, err := s.queries.GetPlatformOutfitSynchronizedAt(ctx, db.GetPlatformOutfitSynchronizedAtParams{
		Platform: string(platform),
		OutfitID: string(outfitId),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return time, shared.ErrNotFound
	}
	return time, err
}

func (s *Storage) TrackingChannelsForCharacter(
	ctx context.Context,
	platform ps2_platforms.Platform,
	characterId ps2.CharacterId,
	outfitId ps2.OutfitId,
) ([]discord.Channel, error) {
	rows, err := s.queries.ListPlatformTrackingChannelsForCharacter(ctx, db.ListPlatformTrackingChannelsForCharacterParams{
		Platform:    string(platform),
		CharacterID: string(characterId),
		OutfitID:    string(outfitId),
	})
	if err != nil {
		return nil, err
	}
	channels := make([]discord.Channel, 0, len(rows))
	for _, row := range rows {
		channels = append(channels, s.dtoToChannel(ctx, row))
	}
	return channels, nil
}

func (s *Storage) TrackingChannelsForOutfit(
	ctx context.Context,
	platform ps2_platforms.Platform,
	outfitId ps2.OutfitId,
) ([]discord.Channel, error) {
	rows, err := s.queries.ListPlatformTrackingChannelsForOutfit(ctx, db.ListPlatformTrackingChannelsForOutfitParams{
		Platform: string(platform),
		OutfitID: string(outfitId),
	})
	if err != nil {
		return nil, err
	}
	channels := make([]discord.Channel, 0, len(rows))
	for _, row := range rows {
		channels = append(channels, s.dtoToChannel(ctx, row))
	}
	return channels, nil
}

func (s *Storage) TrackingOutfitIdsForPlatform(ctx context.Context, channelId discord.ChannelId, platform ps2_platforms.Platform) ([]ps2.OutfitId, error) {
	list, err := s.queries.ListChannelOutfitIdsForPlatform(ctx, db.ListChannelOutfitIdsForPlatformParams{
		ChannelID: string(channelId),
		Platform:  string(platform),
	})
	if err != nil {
		return nil, err
	}
	ids := make([]ps2.OutfitId, 0, len(list))
	for _, id := range list {
		ids = append(ids, ps2.OutfitId(id))
	}
	return ids, nil
}

func (s *Storage) TrackingCharacterIdsForPlatform(ctx context.Context, channelId discord.ChannelId, platform ps2_platforms.Platform) ([]ps2.CharacterId, error) {
	list, err := s.queries.ListChannelCharacterIdsForPlatform(ctx, db.ListChannelCharacterIdsForPlatformParams{
		ChannelID: string(channelId),
		Platform:  string(platform),
	})
	if err != nil {
		return nil, err
	}
	ids := make([]ps2.CharacterId, 0, len(list))
	for _, id := range list {
		ids = append(ids, ps2.CharacterId(id))
	}
	return ids, nil
}

func (s *Storage) AllTrackableCharacterIdsWithDuplicationsForPlatform(ctx context.Context, platform ps2_platforms.Platform) ([]ps2.CharacterId, error) {
	list, err := s.queries.ListTrackableCharacterIdsWithDuplicationForPlatform(ctx, string(platform))
	if err != nil {
		return nil, err
	}
	ids := make([]ps2.CharacterId, 0, len(list))
	for _, id := range list {
		ids = append(ids, ps2.CharacterId(id))
	}
	return ids, nil
}

func (s *Storage) AllTrackableOutfitIdsWithDuplicationsForPlatform(ctx context.Context, platform ps2_platforms.Platform) ([]ps2.OutfitId, error) {
	list, err := s.queries.ListTrackableOutfitIdsWithDuplicationForPlatform(ctx, string(platform))
	if err != nil {
		return nil, err
	}
	ids := make([]ps2.OutfitId, 0, len(list))
	for _, id := range list {
		ids = append(ids, ps2.OutfitId(id))
	}
	return ids, nil
}

func (s *Storage) AllUniqueTrackableOutfitIdsForPlatform(ctx context.Context, platform ps2_platforms.Platform) ([]ps2.OutfitId, error) {
	list, err := s.queries.ListUniqueTrackableOutfitIdsForPlatform(ctx, string(platform))
	if err != nil {
		return nil, err
	}
	ids := make([]ps2.OutfitId, 0, len(list))
	for _, id := range list {
		ids = append(ids, ps2.OutfitId(id))
	}
	return ids, nil
}

func (s *Storage) ChannelTrackablePlatforms(ctx context.Context, channelId discord.ChannelId) ([]ps2_platforms.Platform, error) {
	ps, err := s.queries.ListChannelTrackablePlatforms(ctx, string(channelId))
	if err != nil {
		return nil, err
	}
	platforms := make([]ps2_platforms.Platform, 0, len(ps))
	for _, p := range ps {
		platforms = append(platforms, ps2_platforms.Platform(p))
	}
	return platforms, nil
}

func (s *Storage) OutfitMembers(ctx context.Context, platform ps2_platforms.Platform, outfitId ps2.OutfitId) ([]ps2.CharacterId, error) {
	list, err := s.queries.ListPlatformOutfitMembers(ctx, db.ListPlatformOutfitMembersParams{
		Platform: string(platform),
		OutfitID: string(outfitId),
	})
	if err != nil {
		return nil, err
	}
	ids := make([]ps2.CharacterId, 0, len(list))
	for _, id := range list {
		ids = append(ids, ps2.CharacterId(id))
	}
	return ids, nil
}

func (s *Storage) Outfit(ctx context.Context, platform ps2_platforms.Platform, outfitId ps2.OutfitId) (ps2.Outfit, error) {
	outfit, err := s.queries.GetPlatformOutfit(ctx, db.GetPlatformOutfitParams{
		Platform: string(platform),
		OutfitID: string(outfitId),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ps2.Outfit{}, shared.ErrNotFound
		}
		return ps2.Outfit{}, err
	}
	return ps2.Outfit{
		Id:       ps2.OutfitId(outfit.OutfitID),
		Name:     outfit.OutfitName,
		Tag:      outfit.OutfitTag,
		Platform: platform,
	}, nil
}

func (s *Storage) SaveOutfit(ctx context.Context, outfit ps2.Outfit) error {
	return s.queries.InsertOutfit(ctx, db.InsertOutfitParams{
		Platform:   string(outfit.Platform),
		OutfitID:   string(outfit.Id),
		OutfitName: outfit.Name,
		OutfitTag:  outfit.Tag,
	})
}

func (s *Storage) Outfits(ctx context.Context, platform ps2_platforms.Platform, outfitIds []ps2.OutfitId) ([]ps2.Outfit, error) {
	ids := make([]string, 0, len(outfitIds))
	for _, id := range outfitIds {
		ids = append(ids, string(id))
	}
	list, err := s.queries.ListPlatformOutfits(ctx, db.ListPlatformOutfitsParams{
		Platform:  string(platform),
		OutfitIds: ids,
	})
	if err != nil {
		return nil, err
	}
	outfits := make([]ps2.Outfit, 0, len(list))
	for _, outfit := range list {
		outfits = append(outfits, ps2.Outfit{
			Id:       ps2.OutfitId(outfit.OutfitID),
			Name:     outfit.OutfitName,
			Tag:      outfit.OutfitTag,
			Platform: platform,
		})
	}
	return outfits, nil
}

func (s *Storage) Facility(ctx context.Context, facilityId ps2.FacilityId) (ps2.Facility, error) {
	facility, err := s.queries.GetFacility(ctx, string(facilityId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ps2.Facility{}, shared.ErrNotFound
		}
		return ps2.Facility{}, err
	}
	return ps2.Facility{
		Id:     ps2.FacilityId(facility.FacilityID),
		Name:   facility.FacilityName,
		Type:   facility.FacilityType,
		ZoneId: ps2.ZoneId(facility.ZoneID),
	}, nil
}

func (s *Storage) SaveFacility(ctx context.Context, facility ps2.Facility) error {
	return s.queries.InsertFacility(ctx, db.InsertFacilityParams{
		FacilityID:   string(facility.Id),
		FacilityName: facility.Name,
		FacilityType: facility.Type,
		ZoneID:       string(facility.ZoneId),
	})
}

func (s *Storage) Channel(
	ctx context.Context,
	channelId discord.ChannelId,
) (discord.Channel, error) {
	c, err := s.queries.GetChannel(ctx, string(channelId))
	if err == nil {
		return s.dtoToChannel(ctx, c), nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return discord.NewDefaultChannel(channelId), nil
	}
	return discord.Channel{}, err
}

func (s *Storage) SaveChannelLanguage(
	ctx context.Context,
	channelId discord.ChannelId,
	locale language.Tag,
) error {
	err := s.queries.UpsertChannelLanguage(ctx, db.UpsertChannelLanguageParams{
		ChannelID: string(channelId),
		Locale:    locale.String(),
	})
	return s.publish(err, storage.ChannelLanguageSaved{
		ChannelId: channelId,
		Language:  locale,
	})
}

func (s *Storage) SaveChannelCharacterNotifications(
	ctx context.Context,
	channelId discord.ChannelId,
	enabled bool,
) error {
	err := s.queries.UpsertChannelCharacterNotifications(ctx, db.UpsertChannelCharacterNotificationsParams{
		ChannelID:              string(channelId),
		CharacterNotifications: enabled,
	})
	return s.publish(err, storage.ChannelCharacterNotificationsSaved{
		ChannelId: channelId,
		Enabled:   enabled,
	})
}

func (s *Storage) SaveChannelOutfitNotifications(
	ctx context.Context,
	channelId discord.ChannelId,
	enabled bool,
) error {
	err := s.queries.UpsertChannelOutfitNotifications(ctx, db.UpsertChannelOutfitNotificationsParams{
		ChannelID:           string(channelId),
		OutfitNotifications: enabled,
	})
	return s.publish(err, storage.ChannelOutfitNotificationsSaved{
		ChannelId: channelId,
		Enabled:   enabled,
	})
}

func (s *Storage) SaveChannelTitleUpdates(
	ctx context.Context,
	channelId discord.ChannelId,
	enabled bool,
) error {
	err := s.queries.UpsertChannelTitleUpdates(ctx, db.UpsertChannelTitleUpdatesParams{
		ChannelID:    string(channelId),
		TitleUpdates: enabled,
	})
	return s.publish(err, storage.ChannelTitleUpdatesSaved{
		ChannelId: channelId,
		Enabled:   enabled,
	})
}

func (s *Storage) SaveChannelDefaultTimezone(
	ctx context.Context,
	channelId discord.ChannelId,
	loc *time.Location,
) error {
	err := s.queries.UpsertChannelDefaultTimezone(ctx, db.UpsertChannelDefaultTimezoneParams{
		ChannelID:       string(channelId),
		DefaultTimezone: loc.String(),
	})
	return s.publish(err, storage.ChannelDefaultTimezoneSaved{
		ChannelId: channelId,
		Location:  loc,
	})
}

func (s *Storage) publish(err error, event storage.Event) error {
	if errors.Is(err, sql.ErrNoRows) {
		return shared.ErrNotFound
	}
	if err != nil {
		return err
	}
	s.publisher.Publish(event)
	return nil
}

func (s *Storage) dtoToChannel(ctx context.Context, dto db.Channel) discord.Channel {
	channelId := discord.ChannelId(dto.ChannelID)
	locale, err := language.Parse(dto.Locale)
	if err != nil {
		s.log.Warn(ctx, "failed to parse locale", slog.String("channel_id", string(channelId)), slog.String("locale", dto.Locale), sl.Err(err))
		locale = discord.DEFAULT_LANG_TAG
	}
	loc, err := time.LoadLocation(dto.DefaultTimezone)
	if err != nil {
		s.log.Warn(ctx, "failed to load timezone", slog.String("channel_id", string(channelId)), slog.String("timezone", dto.DefaultTimezone), sl.Err(err))
		loc = time.UTC
	}
	return discord.NewChannel(
		channelId,
		locale,
		dto.CharacterNotifications,
		dto.OutfitNotifications,
		dto.TitleUpdates,
		loc,
	)
}
