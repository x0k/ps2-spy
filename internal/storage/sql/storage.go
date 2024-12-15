package sql_storage

import (
	"context"
	"database/sql"
	"errors"
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
	log         *logger.Logger
	storagePath string
	db          *sql.DB
	queries     *db.Queries
	publisher   pubsub.Publisher[storage.Event]
}

func New(
	log *logger.Logger,
	storagePath string,
	publisher pubsub.Publisher[storage.Event],
) *Storage {
	return &Storage{
		log:         log,
		storagePath: storagePath,
		publisher:   publisher,
		db:          nil,
		queries:     nil,
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

func (s *Storage) SaveChannel(
	ctx context.Context,
	channel discord.Channel,
) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			s.log.Error(ctx, "failed to rollback transaction", sl.Err(err))
		}
	}()
	q := s.queries.WithTx(tx)
	oldChannelDTO, err := q.GetChannel(ctx, string(channel.Id))
	isNoRows := errors.Is(err, sql.ErrNoRows)
	if err != nil && !isNoRows {
		return err
	}
	if err := q.UpsertChannel(ctx, db.UpsertChannelParams{
		ChannelID:              string(channel.Id),
		Locale:                 channel.Locale.String(),
		CharacterNotifications: channel.CharacterNotifications,
		OutfitNotifications:    channel.OutfitNotifications,
		TitleUpdates:           channel.TitleUpdates,
	}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	var oldChannel discord.Channel
	if isNoRows {
		oldChannel = discord.NewDefaultChannel(channel.Id)
	} else {
		oldChannel = s.dtoToChannel(ctx, oldChannelDTO)
	}
	return s.publish(err, storage.ChannelSaved{
		OldChannel: oldChannel,
		NewChannel: channel,
	})
}

func (s *Storage) publish(err error, event storage.Event) error {
	if errors.Is(err, sql.ErrNoRows) {
		return shared.ErrNotFound
	}
	if err != nil {
		return err
	}
	return s.publisher.Publish(event)
}

func (s *Storage) dtoToChannel(ctx context.Context, dto db.Channel) discord.Channel {
	channelId := discord.ChannelId(dto.ChannelID)
	locale, err := language.Parse(dto.Locale)
	if err != nil {
		s.log.Warn(ctx, "failed to parse locale", slog.String("channel_id", string(channelId)), slog.String("locale", dto.Locale), sl.Err(err))
		locale = discord.DEFAULT_LANG_TAG
	}
	return discord.NewChannel(
		channelId,
		locale,
		dto.CharacterNotifications,
		dto.OutfitNotifications,
		dto.TitleUpdates,
	)
}
