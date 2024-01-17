package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/storage"
)

const (
	insertChannelOutfit = iota
	deleteChannelOutfit
	insertChannelCharacter
	deleteChannelCharacter
	selectChannelsByCharacter
	selectChannelOutfitsForPlatform
	selectChannelCharactersForPlatform
	selectAllOutfitsForPlatform
	selectOutfitMembers
	statementsCount
)

var ErrTransactionNotStarted = errors.New("transaction not started")

type Storage struct {
	log         *slog.Logger
	db          *sql.DB
	statements  [statementsCount]*sql.Stmt
	tx          *sql.Tx
	publisher   *storage.Publisher
	eventsBatch []storage.Event
}

func (s *Storage) migrate(ctx context.Context) error {

	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS outfit_members (
	outfit_tag TEXT NOT NULL,
	character_id TEXT NOT NULL,	
	PRIMARY KEY (outfit_tag, character_id)
)
`)
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
	log *slog.Logger,
	storagePath string,
	publisher *storage.Publisher,
) (*Storage, error) {
	const op = "storage.sqlite.New"
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{
		db:  db,
		log: log.With(slog.String("component", "sqlite")),
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
		{
			name: selectChannelsByCharacter,
			stmt: `SELECT channel_id FROM channel_to_character WHERE character_id = ?
				   UNION
				   SELECT channel_id FROM channel_to_outfit WHERE outfit_tag = lower(?)`,
		},
		{selectChannelOutfitsForPlatform, "SELECT outfit_tag FROM channel_to_outfit WHERE channel_id = ? AND platform = ?"},
		{selectChannelCharactersForPlatform, "SELECT character_id FROM channel_to_character WHERE channel_id = ? AND platform = ?"},
		{selectAllOutfitsForPlatform, "SELECT DISTINCT outfit_tag FROM channel_to_outfit WHERE platform = ?"},
		{selectOutfitMembers, "SELECT character_id FROM outfit_members WHERE outfit_tag = lower(?)"},
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

func (s *Storage) Close() error {
	s.log.Info("closing storage")
	const op = "storage.sqlite.Close"
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

func (s *Storage) Begin(ctx context.Context, expectedEventsCount int) (*Storage, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &Storage{
		log:         s.log,
		db:          s.db,
		statements:  s.statements,
		tx:          tx,
		eventsBatch: make([]storage.Event, 0, expectedEventsCount),
	}, nil
}

func (s *Storage) Commit() error {
	if s.tx == nil {
		return ErrTransactionNotStarted
	}
	err := s.tx.Commit()
	if err != nil {
		return err
	}
	for _, event := range s.eventsBatch {
		s.publisher.Publish(event)
	}
	s.eventsBatch = nil
	s.tx = nil
	return nil
}

func (s *Storage) Rollback() error {
	if s.tx == nil {
		return ErrTransactionNotStarted
	}
	s.eventsBatch = nil
	return s.tx.Rollback()
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
			s.log.Error("cannot scan row", sl.Err(err))
			continue
		}
		results = append(results, result)
	}
	return results, rows.Err()
}

func (s *Storage) publish(event storage.Event) {
	if s.tx != nil {
		s.eventsBatch = append(s.eventsBatch, event)
		return
	}
	s.publisher.Publish(event)
}

func (s *Storage) SaveChannelOutfit(ctx context.Context, channelId, platform, outfitID string) error {
	const op = "storage.sqlite.SaveChannelOutfit"
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

func (s *Storage) TrackingChannelIdsForCharacter(ctx context.Context, characterId, outfitTag string) ([]string, error) {
	const op = "storage.sqlite.TrackingChannelIdsForCharacter"
	rows, err := query[string](ctx, s, selectChannelsByCharacter, characterId, outfitTag)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (s *Storage) TrackingOutfitsForPlatform(ctx context.Context, channelId, platform string) ([]string, error) {
	const op = "storage.sqlite.TrackingOutfitsForPlatform"
	rows, err := query[string](ctx, s, selectChannelOutfitsForPlatform, channelId, platform)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (s *Storage) TrackingCharactersForPlatform(ctx context.Context, channelId, platform string) ([]string, error) {
	const op = "storage.sqlite.TrackingCharactersForPlatform"
	rows, err := query[string](ctx, s, selectChannelCharactersForPlatform, channelId, platform)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (s *Storage) AllTrackableOutfitsForPlatform(ctx context.Context, platform string) ([]string, error) {
	const op = "storage.sqlite.AllTrackableOutfitsForPlatform"
	rows, err := query[string](ctx, s, selectAllOutfitsForPlatform, platform)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (s *Storage) OutfitMembers(ctx context.Context, outfitTag string) ([]string, error) {
	const op = "storage.sqlite.OutfitMembers"
	rows, err := query[string](ctx, s, selectOutfitMembers, outfitTag)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}
