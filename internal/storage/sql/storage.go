package sql_storage

import (
	"context"
	"database/sql"
	"errors"
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

	_ "modernc.org/sqlite"
)

type Storage struct {
	name        string
	log         *logger.Logger
	storagePath string
	db          *sql.DB
	queries     *db.Queries
	publisher   pubsub.Publisher[storage.Event]
}

func New(
	name string,
	log *logger.Logger,
	storagePath string,
	publisher pubsub.Publisher[storage.Event],
) *Storage {
	return &Storage{
		name:        name,
		log:         log,
		storagePath: storagePath,
		publisher:   publisher,
		db:          nil,
		queries:     nil,
	}
}

func (s *Storage) Name() string {
	return s.name
}

func (s *Storage) Open(ctx context.Context) error {
	var err error
	s.db, err = sql.Open("sqlite", s.storagePath)
	if err != nil {
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
		if err := tx.Rollback(); err != nil {
			s.log.Error(ctx, "failed to rollback transaction", sl.Err(err))
		}
	}()
	bufferedPublisher := pubsub.NewBufferedPublisher(s.publisher, expectedEventsCount)
	tmp := &Storage{
		log:       s.log,
		db:        s.db,
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

func (s *Storage) SaveChannelOutfit(ctx context.Context, channelId discord.ChannelId, platform ps2_platforms.Platform, outfitId ps2.OutfitId) error {
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

func (s *Storage) DeleteChannelOutfit(ctx context.Context, channelId discord.ChannelId, platform ps2_platforms.Platform, outfitId ps2.OutfitId) error {
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

func (s *Storage) SaveChannelCharacter(ctx context.Context, channelId discord.ChannelId, platform ps2_platforms.Platform, characterId ps2.CharacterId) error {
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

func (s *Storage) DeleteChannelCharacter(ctx context.Context, channelId discord.ChannelId, platform ps2_platforms.Platform, characterId ps2.CharacterId) error {
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

func (s *Storage) SaveOutfitMember(ctx context.Context, platform ps2_platforms.Platform, outfitId ps2.OutfitId, characterId ps2.CharacterId) error {
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

func (s *Storage) DeleteOutfitMember(ctx context.Context, platform ps2_platforms.Platform, outfitId ps2.OutfitId, characterId ps2.CharacterId) error {
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

func (s *Storage) SaveOutfitSynchronizedAt(ctx context.Context, platform ps2_platforms.Platform, outfitId ps2.OutfitId, synchronizedAt time.Time) error {
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

func (s *Storage) OutfitSynchronizedAt(ctx context.Context, platform ps2_platforms.Platform, outfitId ps2.OutfitId) (time.Time, error) {
	return s.queries.GetPlatformOutfitSynchronizedAt(ctx, db.GetPlatformOutfitSynchronizedAtParams{
		Platform: string(platform),
		OutfitID: string(outfitId),
	})
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
		locale := discord.DEFAULT_LOCALE
		if row.Locale.Valid {
			locale = discord.Locale(row.Locale.String)
		}
		channels = append(channels, discord.NewChannel(
			discord.ChannelId(row.ChannelID),
			discord.Locale(locale),
		))
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
	for _, id := range rows {
		locale := discord.DEFAULT_LOCALE
		if id.Locale.Valid {
			locale = discord.Locale(id.Locale.String)
		}
		channels = append(channels, discord.NewChannel(
			discord.ChannelId(id.ChannelID),
			discord.Locale(locale),
		))
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

func (s *Storage) publish(err error, event storage.Event) error {
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return shared.ErrNotFound
		}
		return err
	}
	return s.publisher.Publish(event)
}
