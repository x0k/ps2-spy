package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage"
)

var ErrTransactionNotStarted = errors.New("transaction not started")

const (
	// Simple statements
	insertChannelOutfit = iota
	deleteChannelOutfit
	insertChannelCharacter
	deleteChannelCharacter
	insertOutfitCharacter
	deleteOutfitCharacter
	upsertOutfitSynchronization
	selectOutfitSynchronization
	selectTrackingChannelsForCharacter
	selectTrackingChannelsForOutfit
	selectChannelOutfitsForPlatform
	selectChannelCharactersForPlatform
	selectAllTrackableCharactersWithDuplicationForPlatform
	selectAllTrackableOutfitsWithDuplicationForPlatform
	selectAllUniqueTrackableOutfitsForPlatform
	selectOutfitMembers
	selectOutfit
	insertOutfit
	selectFacility
	insertFacility
	// Dynamic statements
	selectOutfits
	statementsCount
)

var statementNames = [statementsCount]string{
	"insertChannelOutfit",
	"deleteChannelOutfit",
	"insertChannelCharacter",
	"deleteChannelCharacter",
	"insertOutfitCharacter",
	"deleteOutfitCharacter",
	"upsertOutfitSynchronization",
	"selectOutfitSynchronization",
	"selectTrackingChannelsForCharacter",
	"selectTrackingChannelsForOutfit",
	"selectChannelOutfitsForPlatform",
	"selectChannelCharactersForPlatform",
	"selectAllTrackableCharactersWithDuplicationForPlatform",
	"selectAllTrackableOutfitsWithDuplicationForPlatform",
	"selectAllUniqueTrackableOutfitsForPlatform",
	"selectOutfitMembers",
	"selectOutfit",
	"insertOutfit",
	"selectFacility",
	"insertFacility",
	"selectOutfits",
}

type Storage struct {
	log        *logger.Logger
	db         *sql.DB
	statements [statementsCount]statement
	tx         *sql.Tx
	pub        publisher.Abstract[publisher.Event]
}

func (s *Storage) migrate(ctx context.Context) error {

	// TODO: Normalize schema by extracting (platform, outfit_id) into separate table
	//       Maybe also need to extract (platform, character_id)

	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS outfit_to_character (
	platform TEXT NOT NULL,
	outfit_id TEXT NOT NULL,
	character_id TEXT NOT NULL,	
	PRIMARY KEY (platform, outfit_id, character_id)
);`)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS outfit_synchronization (
	platform TEXT NOT NULL,
	outfit_id TEXT NOT NULL,
	synchronized_at TIMESTAMP NOT NULL,
	PRIMARY KEY (platform, outfit_id)
);`)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS channel_to_outfit (
	channel_id TEXT NOT NULL,
	platform TEXT NOT NULL,
	outfit_id TEXT NOT NULL,
	PRIMARY KEY (channel_id, platform, outfit_id)
);`)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS channel_to_character (
	channel_id TEXT NOT NULL,
	platform TEXT NOT NULL,
	character_id TEXT NOT NULL,
	PRIMARY KEY (channel_id, platform, character_id)
);`)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS outfit (
	platform TEXT NOT NULL,
	outfit_id TEXT NOT NULL,
	outfit_name TEXT NOT NULL,
	outfit_tag TEXT NOT NULL,
	PRIMARY KEY (platform, outfit_id)
);`)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS facility (
	facility_id TEXT PRIMARY KEY NOT NULL,
	facility_name TEXT NOT NULL,
	facility_type TEXT NOT NULL,
	zone_id TEXT NOT NULL
);`)
	return err
}

func New(
	ctx context.Context,
	log *logger.Logger,
	storagePath string,
	pub publisher.Abstract[publisher.Event],
) (*Storage, error) {
	const op = "storage.sqlite.New"
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	// db.SetMaxOpenConns(1)
	return &Storage{
		log: log.With(slog.String("component", "storage.sqlite.Storage")),
		db:  db,
		pub: pub,
	}, nil
}

func (s *Storage) Start(ctx context.Context) error {
	const op = "storage.sqlite.Start"
	if err := s.migrate(ctx); err != nil {
		return fmt.Errorf("%s cannot migrate: %w", op, err)
	}
	staticStatements := [...]struct {
		name int
		stmt string
	}{
		{insertChannelOutfit, "INSERT INTO channel_to_outfit VALUES (?, ?, ?)"},
		{deleteChannelOutfit, "DELETE FROM channel_to_outfit WHERE channel_id = ? AND platform = ? AND outfit_id = ?"},
		{insertChannelCharacter, "INSERT INTO channel_to_character VALUES (?, ?, ?)"},
		{deleteChannelCharacter, "DELETE FROM channel_to_character WHERE channel_id = ? AND platform = ? AND character_id = ?"},
		{insertOutfitCharacter, "INSERT INTO outfit_to_character VALUES (?, ?, ?)"},
		{deleteOutfitCharacter, "DELETE FROM outfit_to_character WHERE platform = ? AND outfit_id = ? AND character_id = ?"},
		{upsertOutfitSynchronization, "INSERT INTO outfit_synchronization VALUES (?, ?, ?) ON CONFLICT(platform, outfit_id) DO UPDATE SET synchronized_at = EXCLUDED.synchronized_at"},
		{selectOutfitSynchronization, "SELECT synchronized_at FROM outfit_synchronization WHERE platform = ? AND outfit_id = ?"},
		{
			name: selectTrackingChannelsForCharacter,
			stmt: `SELECT channel_id FROM channel_to_character WHERE platform = ? AND character_id = ?
				   UNION
				   SELECT channel_id FROM channel_to_outfit WHERE platform = ? AND outfit_id = ?`,
		},
		{selectTrackingChannelsForOutfit, "SELECT channel_id FROM channel_to_outfit WHERE platform = ? AND outfit_id = ?"},
		{selectChannelOutfitsForPlatform, "SELECT outfit_id FROM channel_to_outfit WHERE channel_id = ? AND platform = ?"},
		{selectChannelCharactersForPlatform, "SELECT character_id FROM channel_to_character WHERE channel_id = ? AND platform = ?"},
		{
			name: selectAllTrackableCharactersWithDuplicationForPlatform,
			stmt: `SELECT character_id FROM channel_to_character WHERE platform = ?
				   UNION ALL
				   SELECT character_id
				   FROM channel_to_outfit
				   JOIN outfit_to_character ON channel_to_outfit.outfit_id = outfit_to_character.outfit_id AND channel_to_outfit.platform = outfit_to_character.platform
				   WHERE channel_to_outfit.platform = ?`,
		},
		{selectAllTrackableOutfitsWithDuplicationForPlatform, "SELECT outfit_id FROM channel_to_outfit WHERE platform = ?"},
		{selectAllUniqueTrackableOutfitsForPlatform, "SELECT DISTINCT outfit_id FROM channel_to_outfit WHERE platform = ?"},
		{selectOutfitMembers, "SELECT character_id FROM outfit_to_character WHERE platform = ? AND outfit_id = ?"},
		{selectOutfit, "SELECT * FROM outfit WHERE platform = ? AND outfit_id = ?"},
		{insertOutfit, "INSERT INTO outfit VALUES (?, ?, ?, ?)"},
		{selectFacility, "SELECT * FROM facility WHERE facility_id = ?"},
		{insertFacility, "INSERT INTO facility VALUES (?, ?, ?, ?)"},
	}
	for _, simple := range staticStatements {
		stmt, err := s.db.Prepare(simple.stmt)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		s.statements[simple.name] = staticStatement{stmt}
	}
	dynamicStatements := [...]struct {
		name         int
		queryBuilder func(args ...any) (string, error)
	}{
		{selectOutfits, func(args ...any) (string, error) {
			b := strings.Builder{}
			b.WriteString("SELECT * FROM outfit WHERE platform = ? AND outfit_id IN (")
			b.WriteByte('?')
			for i := 2; i < len(args); i++ {
				b.WriteString(", ?")
			}
			b.WriteByte(')')
			return b.String(), nil
		}},
	}
	for _, dynamic := range dynamicStatements {
		s.statements[dynamic.name] = &dynamicStatement{
			db:           s.db,
			queryBuilder: dynamic.queryBuilder,
		}
	}
	return nil
}

func (s *Storage) Close(ctx context.Context) error {
	const op = "storage.sqlite.Close"
	s.log.Info(ctx, "closing sqlite storage")
	errs := make([]string, 0, statementsCount+1)
	for _, st := range s.statements {
		if err := st.Close(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if err := s.db.Close(); err != nil {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s: %s", op, strings.Join(errs, ", "))
	}
	return nil
}

func (s *Storage) Begin(
	ctx context.Context,
	expectedEventsCount int,
	run func(s *Storage) error,
) error {
	const op = "storage.sqlite.Begin"
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	txPublisher := storage.NewTxPublisher(s.pub, expectedEventsCount)
	tmp := &Storage{
		log:        s.log,
		db:         s.db,
		statements: s.statements,
		pub:        txPublisher,
		tx:         tx,
	}
	err = run(tmp)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			s.log.Error(ctx, "cannot rollback transaction", sl.Err(err))
		}
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return txPublisher.Commit()
}

func (s *Storage) stmt(ctx context.Context, st int) statement {
	stmt := s.statements[st]
	if s.tx == nil {
		return stmt
	}
	return stmt.UseTx(ctx, s.tx)
}

func (s *Storage) exec(ctx context.Context, statement int, args ...any) error {
	s.log.Debug(ctx, "execute", slog.String("statement", statementNames[statement]), slog.Any("args", args))
	if _, err := s.stmt(ctx, statement).ExecContext(ctx, args...); err != nil {
		return err
	}
	return nil
}

func defaultScan[T any](r row) (T, error) {
	var result T
	err := r.Scan(&result)
	return result, err
}

func outfitScan(r row) (ps2.Outfit, error) {
	var outfit ps2.Outfit
	return outfit, r.Scan(&outfit.Platform, &outfit.Id, &outfit.Name, &outfit.Tag)
}

func facilityScan(r row) (ps2.Facility, error) {
	var facility ps2.Facility
	return facility, r.Scan(&facility.Id, &facility.Name, &facility.Type, &facility.ZoneId)
}

func queryRow[T any](
	ctx context.Context,
	s *Storage,
	statement int,
	customScan func(row) (T, error),
	args ...any,
) (T, error) {
	s.log.Debug(ctx, "query row", slog.String("statement", statementNames[statement]), slog.Any("args", args))
	if customScan == nil {
		customScan = defaultScan
	}
	result, err := customScan(
		s.stmt(ctx, statement).QueryRowContext(ctx, args...),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return result, storage.ErrNotFound
	}
	return result, err
}

func query[T any](
	ctx context.Context,
	s *Storage,
	statement int,
	customScan func(row) (T, error),
	args ...any,
) ([]T, error) {
	s.log.Debug(ctx, "query", slog.String("statement", statementNames[statement]), slog.Any("args", args))
	rows, err := s.stmt(ctx, statement).QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if customScan == nil {
		customScan = defaultScan
	}
	results := make([]T, 0)
	for rows.Next() {
		result, err := customScan(rows)
		if err != nil {
			s.log.Error(ctx, "cannot scan row", sl.Err(err))
			continue
		}
		results = append(results, result)
	}
	return results, rows.Err()
}

func (s *Storage) publish(event publisher.Event) {
	s.pub.Publish(event)
}

func (s *Storage) SaveChannelOutfit(ctx context.Context, channelId meta.ChannelId, platform platforms.Platform, outfitId ps2.OutfitId) error {
	const op = "storage.sqlite.SaveChannelOutfit"
	err := s.exec(ctx, insertChannelOutfit, channelId, platform, outfitId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	s.publish(storage.ChannelOutfitSaved{
		ChannelId: channelId,
		Platform:  platform,
		OutfitId:  outfitId,
	})
	return nil
}

func (s *Storage) DeleteChannelOutfit(ctx context.Context, channelId meta.ChannelId, platform platforms.Platform, outfitId ps2.OutfitId) error {
	const op = "storage.sqlite.DeleteChannelOutfit"
	err := s.exec(ctx, deleteChannelOutfit, channelId, platform, outfitId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	s.publish(storage.ChannelOutfitDeleted{
		ChannelId: channelId,
		Platform:  platform,
		OutfitId:  outfitId,
	})
	return nil
}

func (s *Storage) SaveChannelCharacter(ctx context.Context, channelId meta.ChannelId, platform platforms.Platform, characterId ps2.CharacterId) error {
	const op = "storage.sqlite.SaveChannelCharacter"
	err := s.exec(ctx, insertChannelCharacter, channelId, platform, characterId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	s.publish(storage.ChannelCharacterSaved{
		ChannelId:   channelId,
		Platform:    platform,
		CharacterId: characterId,
	})
	return nil
}

func (s *Storage) DeleteChannelCharacter(ctx context.Context, channelId meta.ChannelId, platform platforms.Platform, characterId ps2.CharacterId) error {
	const op = "storage.sqlite.DeleteChannelCharacter"
	err := s.exec(ctx, deleteChannelCharacter, channelId, platform, characterId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	s.publish(storage.ChannelCharacterDeleted{
		ChannelId:   channelId,
		Platform:    platform,
		CharacterId: characterId,
	})
	return nil
}

func (s *Storage) SaveOutfitMember(ctx context.Context, platform platforms.Platform, outfitId ps2.OutfitId, characterId ps2.CharacterId) error {
	const op = "storage.sqlite.SaveOutfitMember"
	err := s.exec(ctx, insertOutfitCharacter, platform, outfitId, characterId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	s.publish(storage.OutfitMemberSaved{
		Platform:    platform,
		OutfitId:    outfitId,
		CharacterId: characterId,
	})
	return nil
}

func (s *Storage) DeleteOutfitMember(ctx context.Context, platform platforms.Platform, outfitId ps2.OutfitId, characterId ps2.CharacterId) error {
	const op = "storage.sqlite.DeleteOutfitMember"
	err := s.exec(ctx, deleteOutfitCharacter, platform, outfitId, characterId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	s.publish(storage.OutfitMemberDeleted{
		Platform:    platform,
		OutfitId:    outfitId,
		CharacterId: characterId,
	})
	return nil
}

func (s *Storage) SaveOutfitSynchronizedAt(ctx context.Context, platform platforms.Platform, outfitId ps2.OutfitId, at time.Time) error {
	const op = "storage.sqlite.SaveOutfitSynchronizedAt"
	err := s.exec(ctx, upsertOutfitSynchronization, platform, outfitId, at)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	s.publish(storage.OutfitSynchronized{
		Platform:       platform,
		OutfitId:       outfitId,
		SynchronizedAt: at,
	})
	return nil
}

func (s *Storage) OutfitSynchronizedAt(ctx context.Context, platform platforms.Platform, outfitId ps2.OutfitId) (time.Time, error) {
	const op = "storage.sqlite.OutfitSynchronizedAt"
	syncAt, err := queryRow[time.Time](ctx, s, selectOutfitSynchronization, nil, platform, outfitId)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s: %w", op, err)
	}
	return syncAt, nil
}

func (s *Storage) TrackingChannelIdsForCharacter(ctx context.Context, platform platforms.Platform, characterId ps2.CharacterId, outfitId ps2.OutfitId) ([]meta.ChannelId, error) {
	const op = "storage.sqlite.TrackingChannelIdsForCharacter"
	rows, err := query[meta.ChannelId](ctx, s, selectTrackingChannelsForCharacter, nil, platform, characterId, platform, outfitId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (s *Storage) TrackingChannelsIdsForOutfit(ctx context.Context, platform platforms.Platform, outfitId ps2.OutfitId) ([]meta.ChannelId, error) {
	const op = "storage.sqlite.TrackingChannelsIdsForOutfit"
	rows, err := query[meta.ChannelId](ctx, s, selectTrackingChannelsForOutfit, nil, platform, outfitId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (s *Storage) TrackingOutfitsForPlatform(ctx context.Context, channelId meta.ChannelId, platform platforms.Platform) ([]ps2.OutfitId, error) {
	const op = "storage.sqlite.TrackingOutfitsForPlatform"
	rows, err := query[ps2.OutfitId](ctx, s, selectChannelOutfitsForPlatform, nil, channelId, platform)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (s *Storage) TrackingCharactersForPlatform(ctx context.Context, channelId meta.ChannelId, platform platforms.Platform) ([]ps2.CharacterId, error) {
	const op = "storage.sqlite.TrackingCharactersForPlatform"
	rows, err := query[ps2.CharacterId](ctx, s, selectChannelCharactersForPlatform, nil, channelId, platform)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (s *Storage) AllTrackableCharactersWithDuplicationsForPlatform(ctx context.Context, platform platforms.Platform) ([]ps2.CharacterId, error) {
	const op = "storage.sqlite.AllTrackableCharactersForPlatform"
	rows, err := query[ps2.CharacterId](ctx, s, selectAllTrackableCharactersWithDuplicationForPlatform, nil, platform, platform)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (s *Storage) AllTrackableOutfitsWithDuplicationsForPlatform(ctx context.Context, platform platforms.Platform) ([]ps2.OutfitId, error) {
	const op = "storage.sqlite.AllTrackableOutfitsWithDuplicationsForPlatform"
	rows, err := query[ps2.OutfitId](ctx, s, selectAllTrackableOutfitsWithDuplicationForPlatform, nil, platform)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (s *Storage) AllUniqueTrackableOutfitsForPlatform(ctx context.Context, platform platforms.Platform) ([]ps2.OutfitId, error) {
	const op = "storage.sqlite.AllUniqueTrackableOutfitsForPlatform"
	rows, err := query[ps2.OutfitId](ctx, s, selectAllUniqueTrackableOutfitsForPlatform, nil, platform)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (s *Storage) OutfitMembers(ctx context.Context, platform platforms.Platform, outfitId ps2.OutfitId) ([]ps2.CharacterId, error) {
	const op = "storage.sqlite.OutfitMembers"
	rows, err := query[ps2.CharacterId](ctx, s, selectOutfitMembers, nil, platform, outfitId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (s *Storage) Outfit(ctx context.Context, platform platforms.Platform, outfitId ps2.OutfitId) (ps2.Outfit, error) {
	const op = "storage.sqlite.Outfit"
	outfit, err := queryRow(ctx, s, selectOutfit, outfitScan, platform, outfitId)
	if err != nil {
		return ps2.Outfit{}, fmt.Errorf("%s: %w", op, err)
	}
	return outfit, nil
}

func (s *Storage) SaveOutfit(ctx context.Context, outfit ps2.Outfit) error {
	const op = "storage.sqlite.SaveOutfit"
	err := s.exec(ctx, insertOutfit, outfit.Platform, outfit.Id, outfit.Name, outfit.Tag)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) Outfits(ctx context.Context, platform platforms.Platform, outfitIds []ps2.OutfitId) ([]ps2.Outfit, error) {
	const op = "storage.sqlite.Outfits"
	args := make([]any, len(outfitIds)+1)
	args[0] = platform
	for i, id := range outfitIds {
		args[i+1] = id
	}
	rows, err := query(ctx, s, selectOutfits, outfitScan, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (s *Storage) Facility(ctx context.Context, facilityId ps2.FacilityId) (ps2.Facility, error) {
	const op = "storage.sqlite.Facility"
	facility, err := queryRow(ctx, s, selectFacility, facilityScan, facilityId)
	if err != nil {
		return ps2.Facility{}, fmt.Errorf("%s: %w", op, err)
	}
	return facility, nil
}

func (s *Storage) SaveFacility(ctx context.Context, f ps2.Facility) error {
	const op = "storage.sqlite.SaveFacility"
	err := s.exec(ctx, insertFacility, f.Id, f.Name, f.Type, f.ZoneId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
