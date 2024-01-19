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
	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/publisher"
	"github.com/x0k/ps2-spy/internal/storage"
)

const (
	insertChannelOutfit = iota
	deleteChannelOutfit
	insertChannelCharacter
	deleteChannelCharacter
	insertOutfitCharacter
	deleteOutfitCharacter
	upsertOutfitSynchronization
	selectOutfitSynchronization
	selectTrackingChannelsForCharacter
	selectChannelOutfitsForPlatform
	selectChannelCharactersForPlatform
	selectAllTrackableCharactersWithDuplicationForPlatform
	selectAllUniqueTrackableOutfitsForPlatform
	selectOutfitMembers
	countOutfitTrackingChannels
	statementsCount
)

var ErrTransactionNotStarted = errors.New("transaction not started")

type Storage struct {
	db         *sql.DB
	statements [statementsCount]*sql.Stmt
	tx         *sql.Tx
	pub        publisher.Abstract[storage.Event]
}

func (s *Storage) migrate(ctx context.Context) error {

	// TODO: Normalize schema by extracting (platform, outfit_tag) into separate table
	//       Maybe also need to extract (platform, character_id)

	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS outfit_to_character (
	platform TEXT NOT NULL,
	outfit_tag TEXT NOT NULL,
	character_id TEXT NOT NULL,	
	PRIMARY KEY (platform, outfit_tag, character_id)
);`)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS outfit_synchronization (
	platform TEXT NOT NULL,
	outfit_tag TEXT NOT NULL,
	synchronized_at TIMESTAMP NOT NULL,
	PRIMARY KEY (platform, outfit_tag)
);`)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS channel_to_outfit (
	channel_id TEXT NOT NULL,
	platform TEXT NOT NULL,
	outfit_tag TEXT NOT NULL,
	PRIMARY KEY (channel_id, platform, outfit_tag)
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
	return err
}

func New(
	ctx context.Context,
	storagePath string,
	pub publisher.Abstract[storage.Event],
) (*Storage, error) {
	const op = "storage.sqlite.New"
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	// db.SetMaxOpenConns(1)
	return &Storage{
		db:  db,
		pub: pub,
	}, nil
}

func (s *Storage) Start(ctx context.Context) error {
	const op = "storage.sqlite.Start"
	if err := s.migrate(ctx); err != nil {
		return fmt.Errorf("%s cannot migrate: %w", op, err)
	}
	rawStatements := [statementsCount]struct {
		name int
		stmt string
	}{
		{insertChannelOutfit, "INSERT INTO channel_to_outfit VALUES (?, ?, lower(?))"},
		{deleteChannelOutfit, "DELETE FROM channel_to_outfit WHERE channel_id = ? AND platform = ? AND outfit_tag = lower(?)"},
		{insertChannelCharacter, "INSERT INTO channel_to_character VALUES (?, ?, ?)"},
		{deleteChannelCharacter, "DELETE FROM channel_to_character WHERE channel_id = ? AND platform = ? AND character_id = ?"},
		{insertOutfitCharacter, "INSERT INTO outfit_to_character VALUES (?, lower(?), ?)"},
		{deleteOutfitCharacter, "DELETE FROM outfit_to_character WHERE platform = ? AND outfit_tag = lower(?) AND character_id = ?"},
		{upsertOutfitSynchronization, "INSERT INTO outfit_synchronization VALUES (?, lower(?), ?) ON CONFLICT(platform, outfit_tag) DO UPDATE SET synchronized_at = EXCLUDED.synchronized_at"},
		{selectOutfitSynchronization, "SELECT synchronized_at FROM outfit_synchronization WHERE platform = ? AND outfit_tag = lower(?)"},
		{
			name: selectTrackingChannelsForCharacter,
			stmt: `SELECT channel_id FROM channel_to_character WHERE platform = ? AND character_id = ?
				   UNION
				   SELECT channel_id FROM channel_to_outfit WHERE platform = ? AND outfit_tag = lower(?)`,
		},
		{selectChannelOutfitsForPlatform, "SELECT outfit_tag FROM channel_to_outfit WHERE channel_id = ? AND platform = ?"},
		{selectChannelCharactersForPlatform, "SELECT character_id FROM channel_to_character WHERE channel_id = ? AND platform = ?"},
		{
			name: selectAllTrackableCharactersWithDuplicationForPlatform,
			stmt: `SELECT character_id FROM channel_to_character WHERE platform = ?
				   UNION ALL
				   SELECT character_id
				   FROM channel_to_outfit
				   JOIN outfit_to_character ON channel_to_outfit.outfit_tag = outfit_to_character.outfit_tag AND channel_to_outfit.platform = outfit_to_character.platform
				   WHERE channel_to_outfit.platform = ?`,
		},
		{selectAllUniqueTrackableOutfitsForPlatform, "SELECT DISTINCT outfit_tag FROM channel_to_outfit WHERE platform = ?"},
		{selectOutfitMembers, "SELECT character_id FROM outfit_to_character WHERE platform = ? AND outfit_tag = lower(?)"},
		{countOutfitTrackingChannels, "SELECT COUNT(*) FROM channel_to_outfit WHERE platform = ? AND outfit_tag = lower(?)"},
	}
	for _, raw := range rawStatements {
		stmt, err := s.db.Prepare(raw.stmt)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		s.statements[raw.name] = stmt
	}
	return nil
}

func (s *Storage) Close(ctx context.Context) error {
	const op = "storage.sqlite.Close"
	infra.OpLogger(ctx, op).Info("closing sqlite storage")
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
	log := infra.OpLogger(ctx, op)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	txPublisher := storage.NewTxPublisher(s.pub, expectedEventsCount)
	tmp := &Storage{
		db:         s.db,
		statements: s.statements,
		pub:        txPublisher,
		tx:         tx,
	}
	err = run(tmp)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Error("cannot rollback transaction", sl.Err(err))
		}
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return txPublisher.Commit()
}

func (s *Storage) stmt(ctx context.Context, st int) *sql.Stmt {
	if s.tx != nil {
		return s.tx.StmtContext(ctx, s.statements[st])
	}
	return s.statements[st]
}

func (s *Storage) exec(ctx context.Context, statement int, args ...any) error {
	if _, err := s.stmt(ctx, statement).ExecContext(ctx, args...); err != nil {
		return err
	}
	return nil
}

func (s *Storage) queryRow(ctx context.Context, result any, statement int, args ...any) error {
	if err := s.stmt(ctx, statement).QueryRowContext(ctx, args...).Scan(result); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrNotFound
		}
		return err
	}
	return nil
}

func query[T any](ctx context.Context, s *Storage, statement int, args ...any) ([]T, error) {
	const op = "storage.sqlite.query"
	log := infra.OpLogger(ctx, op)
	rows, err := s.stmt(ctx, statement).QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]T, 0)
	for rows.Next() {
		var result T
		err = rows.Scan(&result)
		if err != nil {
			log.Error("cannot scan row", sl.Err(err))
			continue
		}
		results = append(results, result)
	}
	return results, rows.Err()
}

func (s *Storage) publish(event storage.Event) {
	s.pub.Publish(event)
}

func (s *Storage) SaveChannelOutfit(ctx context.Context, channelId, platform, outfitID string) error {
	const op = "storage.sqlite.SaveChannelOutfit"
	infra.OpLogger(ctx, op).Debug("params", slog.String("channelId", channelId), slog.String("platform", platform), slog.String("outfitID", outfitID))
	err := s.exec(ctx, insertChannelOutfit, channelId, platform, outfitID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	s.publish(storage.ChannelOutfitSaved{
		ChannelId: channelId,
		Platform:  platform,
		OutfitId:  outfitID,
	})
	return nil
}

func (s *Storage) DeleteChannelOutfit(ctx context.Context, channelId, platform, outfitID string) error {
	const op = "storage.sqlite.DeleteChannelOutfit"
	infra.OpLogger(ctx, op).Debug("params", slog.String("channelId", channelId), slog.String("platform", platform), slog.String("outfitID", outfitID))
	err := s.exec(ctx, deleteChannelOutfit, channelId, platform, outfitID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	s.publish(storage.ChannelOutfitDeleted{
		ChannelId: channelId,
		Platform:  platform,
		OutfitId:  outfitID,
	})
	return nil
}

func (s *Storage) SaveChannelCharacter(ctx context.Context, channelId, platform, characterID string) error {
	const op = "storage.sqlite.SaveChannelCharacter"
	infra.OpLogger(ctx, op).Debug("params", slog.String("channelId", channelId), slog.String("platform", platform), slog.String("characterID", characterID))
	err := s.exec(ctx, insertChannelCharacter, channelId, platform, characterID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	s.publish(storage.ChannelCharacterSaved{
		ChannelId:   channelId,
		Platform:    platform,
		CharacterId: characterID,
	})
	return nil
}

func (s *Storage) DeleteChannelCharacter(ctx context.Context, channelId, platform, characterID string) error {
	const op = "storage.sqlite.DeleteChannelCharacter"
	infra.OpLogger(ctx, op).Debug("params", slog.String("channelId", channelId), slog.String("platform", platform), slog.String("characterID", characterID))
	err := s.exec(ctx, deleteChannelCharacter, channelId, platform, characterID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	s.publish(storage.ChannelCharacterDeleted{
		ChannelId:   channelId,
		Platform:    platform,
		CharacterId: characterID,
	})
	return nil
}

func (s *Storage) SaveOutfitMember(ctx context.Context, platform, outfitTag, characterId string) error {
	const op = "storage.sqlite.SaveOutfitMember"
	infra.OpLogger(ctx, op).Debug("params", slog.String("platform", platform), slog.String("outfitTag", outfitTag), slog.String("characterId", characterId))
	err := s.exec(ctx, insertOutfitCharacter, platform, outfitTag, characterId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	s.publish(storage.OutfitMemberSaved{
		Platform:    platform,
		OutfitTag:   outfitTag,
		CharacterId: characterId,
	})
	return nil
}

func (s *Storage) DeleteOutfitMember(ctx context.Context, platform, outfitTag, characterId string) error {
	const op = "storage.sqlite.DeleteOutfitMember"
	infra.OpLogger(ctx, op).Debug("params", slog.String("platform", platform), slog.String("outfitTag", outfitTag), slog.String("characterId", characterId))
	err := s.exec(ctx, deleteOutfitCharacter, platform, outfitTag, characterId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	s.publish(storage.OutfitMemberDeleted{
		Platform:    platform,
		OutfitTag:   outfitTag,
		CharacterId: characterId,
	})
	return nil
}

func (s *Storage) SaveOutfitSynchronizedAt(ctx context.Context, platform, outfitTag string, at time.Time) error {
	const op = "storage.sqlite.SaveOutfitSynchronizedAt"
	infra.OpLogger(ctx, op).Debug("params", slog.String("platform", platform), slog.String("outfitTag", outfitTag), slog.Time("at", at))
	err := s.exec(ctx, upsertOutfitSynchronization, platform, outfitTag, at)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	s.publish(storage.OutfitSynchronized{
		Platform:       platform,
		OutfitTag:      outfitTag,
		SynchronizedAt: at,
	})
	return nil
}

func (s *Storage) OutfitSynchronizedAt(ctx context.Context, platform, outfitTag string) (time.Time, error) {
	const op = "storage.sqlite.OutfitSynchronizedAt"
	infra.OpLogger(ctx, op).Debug("params", slog.String("platform", platform), slog.String("outfitTag", outfitTag))
	var syncAt time.Time
	err := s.queryRow(ctx, &syncAt, selectOutfitSynchronization, platform, outfitTag)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s: %w", op, err)
	}
	return syncAt, nil
}

func (s *Storage) TrackingChannelIdsForCharacter(ctx context.Context, platform, characterId, outfitTag string) ([]string, error) {
	const op = "storage.sqlite.TrackingChannelIdsForCharacter"
	infra.OpLogger(ctx, op).Debug("params", slog.String("platform", platform), slog.String("characterId", characterId), slog.String("outfitTag", outfitTag))
	rows, err := query[string](ctx, s, selectTrackingChannelsForCharacter, platform, characterId, platform, outfitTag)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (s *Storage) TrackingOutfitsForPlatform(ctx context.Context, channelId, platform string) ([]string, error) {
	const op = "storage.sqlite.TrackingOutfitsForPlatform"
	infra.OpLogger(ctx, op).Debug("params", slog.String("channelId", channelId), slog.String("platform", platform))
	rows, err := query[string](ctx, s, selectChannelOutfitsForPlatform, channelId, platform)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (s *Storage) TrackingCharactersForPlatform(ctx context.Context, channelId, platform string) ([]string, error) {
	const op = "storage.sqlite.TrackingCharactersForPlatform"
	infra.OpLogger(ctx, op).Debug("params", slog.String("channelId", channelId), slog.String("platform", platform))
	rows, err := query[string](ctx, s, selectChannelCharactersForPlatform, channelId, platform)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (s *Storage) AllTrackableCharactersWithDuplicationsForPlatform(ctx context.Context, platform string) ([]string, error) {
	const op = "storage.sqlite.AllTrackableCharactersForPlatform"
	infra.OpLogger(ctx, op).Debug("params", slog.String("platform", platform))
	rows, err := query[string](ctx, s, selectAllTrackableCharactersWithDuplicationForPlatform, platform, platform)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (s *Storage) AllUniqueTrackableOutfitsForPlatform(ctx context.Context, platform string) ([]string, error) {
	const op = "storage.sqlite.AllTrackableOutfitsForPlatform"
	infra.OpLogger(ctx, op).Debug("params", slog.String("platform", platform))
	rows, err := query[string](ctx, s, selectAllUniqueTrackableOutfitsForPlatform, platform)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (s *Storage) OutfitMembers(ctx context.Context, platform, outfitTag string) ([]string, error) {
	const op = "storage.sqlite.OutfitMembers"
	infra.OpLogger(ctx, op).Debug("params", slog.String("platform", platform), slog.String("outfitTag", outfitTag))
	rows, err := query[string](ctx, s, selectOutfitMembers, platform, outfitTag)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (s *Storage) CountOutfitTrackingChannels(ctx context.Context, platform, outfitTag string) (int, error) {
	const op = "storage.sqlite.CountOutfitTrackingChannels"
	infra.OpLogger(ctx, op).Debug("params", slog.String("platform", platform), slog.String("outfitTag", outfitTag))
	var count int
	err := s.queryRow(ctx, &count, countOutfitTrackingChannels, platform, outfitTag)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return count, nil
}
