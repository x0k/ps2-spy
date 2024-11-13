package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/x0k/ps2-spy/internal/event"
	"github.com/x0k/ps2-spy/internal/lib/db"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage"
)

type SqliteStorage struct {
	log       *logger.Logger
	sqlite    *sql.DB
	queries   *db.Queries
	publisher event.Publisher
}

func New(ctx context.Context, log *logger.Logger, storagePath string, publisher event.Publisher) (*SqliteStorage, error) {
	const op = "storage.sqlite.New"
	sqlite, err := sql.Open("sqlite", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	queries, err := db.Prepare(ctx, sqlite)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &SqliteStorage{
		log:       log.With(sl.Component("storage.sqlite.Storage")),
		sqlite:    sqlite,
		queries:   queries,
		publisher: publisher,
	}, nil
}

func (s *SqliteStorage) Close() error {
	return errors.Join(
		s.queries.Close(),
		s.sqlite.Close(),
	)
}

func (s *SqliteStorage) Begin(
	ctx context.Context,
	expectedEventsCount int,
	run func(s *SqliteStorage) error,
) error {
	tx, err := s.sqlite.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			s.log.Error(ctx, "failed to rollback transaction", sl.Err(err))
		}
	}()
	bufferedPublisher := event.NewBufferedPublisher(s.publisher, expectedEventsCount)
	tmp := &SqliteStorage{
		log:       s.log,
		sqlite:    s.sqlite,
		queries:   s.queries.WithTx(tx),
		publisher: bufferedPublisher,
	}
	err = run(tmp)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return bufferedPublisher.Flush()
}

func (s *SqliteStorage) SaveChannelOutfit(ctx context.Context, channelId meta.ChannelId, platform platforms.Platform, outfitId ps2.OutfitId) error {
	err := s.queries.InsertChannelOutfit(ctx, db.InsertChannelOutfitParams{
		ChannelID: string(channelId),
		OutfitID:  string(outfitId),
		Platform:  string(platform),
	})
	return s.publish(err, storage.ChannelOutfitSaved{
		ChannelId: channelId,
		Platform:  platform,
		OutfitId:  outfitId,
	})
}

func (s *SqliteStorage) DeleteChannelOutfit(ctx context.Context, channelId meta.ChannelId, platform platforms.Platform, outfitId ps2.OutfitId) error {
	err := s.queries.DeleteChannelOutfit(ctx, db.DeleteChannelOutfitParams{
		ChannelID: string(channelId),
		OutfitID:  string(outfitId),
		Platform:  string(platform),
	})
	return s.publish(err, storage.ChannelOutfitDeleted{
		ChannelId: channelId,
		Platform:  platform,
		OutfitId:  outfitId,
	})
}

func (s *SqliteStorage) SaveChannelCharacter(ctx context.Context, channelId meta.ChannelId, platform platforms.Platform, characterId ps2.CharacterId) error {
	err := s.queries.InsertChannelCharacter(ctx, db.InsertChannelCharacterParams{
		ChannelID:   string(channelId),
		CharacterID: string(characterId),
		Platform:    string(platform),
	})
	return s.publish(err, storage.ChannelCharacterSaved{
		ChannelId:   channelId,
		Platform:    platform,
		CharacterId: characterId,
	})
}

func (s *SqliteStorage) DeleteChannelCharacter(ctx context.Context, channelId meta.ChannelId, platform platforms.Platform, characterId ps2.CharacterId) error {
	err := s.queries.DeleteChannelCharacter(ctx, db.DeleteChannelCharacterParams{
		ChannelID:   string(channelId),
		CharacterID: string(characterId),
		Platform:    string(platform),
	})
	return s.publish(err, storage.ChannelCharacterDeleted{
		ChannelId:   channelId,
		Platform:    platform,
		CharacterId: characterId,
	})
}

func (s *SqliteStorage) SaveOutfitMember(ctx context.Context, platform platforms.Platform, outfitId ps2.OutfitId, characterId ps2.CharacterId) error {
	err := s.queries.InsertOutfitMember(ctx, db.InsertOutfitMemberParams{
		OutfitID:    string(outfitId),
		CharacterID: string(characterId),
		Platform:    string(platform),
	})
	return s.publish(err, storage.OutfitMemberSaved{
		Platform:    platform,
		OutfitId:    outfitId,
		CharacterId: characterId,
	})
}

func (s *SqliteStorage) DeleteOutfitMember(ctx context.Context, platform platforms.Platform, outfitId ps2.OutfitId, characterId ps2.CharacterId) error {
	err := s.queries.DeleteOutfitMember(ctx, db.DeleteOutfitMemberParams{
		OutfitID:    string(outfitId),
		CharacterID: string(characterId),
		Platform:    string(platform),
	})
	return s.publish(err, storage.OutfitMemberDeleted{
		Platform:    platform,
		OutfitId:    outfitId,
		CharacterId: characterId,
	})
}

func (s *SqliteStorage) SaveOutfitSynchronizedAt(ctx context.Context, platform platforms.Platform, outfitId ps2.OutfitId, synchronizedAt time.Time) error {
	err := s.queries.UpsertPlatformOutfitSynchronizedAt(ctx, db.UpsertPlatformOutfitSynchronizedAtParams{
		Platform:       string(platform),
		OutfitID:       string(outfitId),
		SynchronizedAt: synchronizedAt,
	})
	return s.publish(err, storage.OutfitSynchronized{
		Platform:       platform,
		OutfitId:       outfitId,
		SynchronizedAt: synchronizedAt,
	})
}

func (s *SqliteStorage) OutfitSynchronizedAt(ctx context.Context, platform platforms.Platform, outfitId ps2.OutfitId) (time.Time, error) {
	return s.queries.GetPlatformOutfitSynchronizedAt(ctx, db.GetPlatformOutfitSynchronizedAtParams{
		Platform: string(platform),
		OutfitID: string(outfitId),
	})
}

func (s *SqliteStorage) TrackingChannelIdsForCharacter(ctx context.Context, platform platforms.Platform, characterId ps2.CharacterId, outfitId ps2.OutfitId) ([]meta.ChannelId, error) {
	list, err := s.queries.ListPlatformTrackingChannelIdsForCharacter(ctx, db.ListPlatformTrackingChannelIdsForCharacterParams{
		Platform:    string(platform),
		CharacterID: string(characterId),
		OutfitID:    string(outfitId),
	})
	if err != nil {
		return nil, err
	}
	ids := make([]meta.ChannelId, 0, len(list))
	for _, id := range list {
		ids = append(ids, meta.ChannelId(id))
	}
	return ids, nil
}

func (s *SqliteStorage) TrackingChannelIdsForOutfit(ctx context.Context, platform platforms.Platform, outfitId ps2.OutfitId) ([]meta.ChannelId, error) {
	list, err := s.queries.ListPlatformTrackingChannelIdsForOutfit(ctx, db.ListPlatformTrackingChannelIdsForOutfitParams{
		Platform: string(platform),
		OutfitID: string(outfitId),
	})
	if err != nil {
		return nil, err
	}
	ids := make([]meta.ChannelId, 0, len(list))
	for _, id := range list {
		ids = append(ids, meta.ChannelId(id))
	}
	return ids, nil
}

func (s *SqliteStorage) TrackingOutfitIdsForPlatform(ctx context.Context, channelId meta.ChannelId, platform platforms.Platform) ([]ps2.OutfitId, error) {
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

func (s *SqliteStorage) TrackingCharacterIdsForPlatform(ctx context.Context, channelId meta.ChannelId, platform platforms.Platform) ([]ps2.CharacterId, error) {
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

func (s *SqliteStorage) AllTrackableCharacterIdsWithDuplicationsForPlatform(ctx context.Context, platform platforms.Platform) ([]ps2.CharacterId, error) {
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

func (s *SqliteStorage) AllTrackableOutfitIdsWithDuplicationsForPlatform(ctx context.Context, platform platforms.Platform) ([]ps2.OutfitId, error) {
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

func (s *SqliteStorage) AllUniqueTrackableOutfitIdsForPlatform(ctx context.Context, platform platforms.Platform) ([]ps2.OutfitId, error) {
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

func (s *SqliteStorage) OutfitMembers(ctx context.Context, platform platforms.Platform, outfitId ps2.OutfitId) ([]ps2.CharacterId, error) {
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

func (s *SqliteStorage) Outfit(ctx context.Context, platform platforms.Platform, outfitId ps2.OutfitId) (ps2.Outfit, error) {
	outfit, err := s.queries.GetPlatformOutfit(ctx, db.GetPlatformOutfitParams{
		Platform: string(platform),
		OutfitID: string(outfitId),
	})
	if err != nil {
		return ps2.Outfit{}, err
	}
	return ps2.Outfit{
		Id:       ps2.OutfitId(outfit.OutfitID),
		Name:     outfit.OutfitName,
		Tag:      outfit.OutfitTag,
		Platform: platform,
	}, nil
}

func (s *SqliteStorage) SaveOutfit(ctx context.Context, outfit ps2.Outfit) error {
	return s.queries.InsertOutfit(ctx, db.InsertOutfitParams{
		Platform:   string(outfit.Platform),
		OutfitID:   string(outfit.Id),
		OutfitName: outfit.Name,
		OutfitTag:  outfit.Tag,
	})
}

func (s *SqliteStorage) Outfits(ctx context.Context, platform platforms.Platform, outfitIds []ps2.OutfitId) ([]ps2.Outfit, error) {
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

func (s *SqliteStorage) Facility(ctx context.Context, facilityId ps2.FacilityId) (ps2.Facility, error) {
	facility, err := s.queries.GetFacility(ctx, string(facilityId))
	if err != nil {
		return ps2.Facility{}, err
	}
	return ps2.Facility{
		Id:     ps2.FacilityId(facility.FacilityID),
		Name:   facility.FacilityName,
		Type:   facility.FacilityType,
		ZoneId: ps2.ZoneId(facility.ZoneID),
	}, nil
}

func (s *SqliteStorage) SaveFacility(ctx context.Context, facility ps2.Facility) error {
	return s.queries.InsertFacility(ctx, db.InsertFacilityParams{
		FacilityID:   string(facility.Id),
		FacilityName: facility.Name,
		FacilityType: facility.Type,
		ZoneID:       string(facility.ZoneId),
	})
}

func (s *SqliteStorage) publish(err error, event event.Event) error {
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrNotFound
		}
		return err
	}
	return s.publisher.Publish(event)
}
