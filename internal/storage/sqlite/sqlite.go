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
	selectOutfitsForPlatform
	selectCharactersForPlatform
	statementsCount
)

var ErrTransactionNotStarted = errors.New("transaction not started")

type Storage struct {
	log        *slog.Logger
	db         *sql.DB
	statements [statementsCount]*sql.Stmt
	tx         *sql.Tx
}

func (s *Storage) migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS channel_to_outfit (
	channel_id TEXT NOT NULL,
	platform_id TEXT NOT NULL,
	outfit_tag TEXT NOT NULL,
	PRIMARY KEY (channel_id, platform_id, outfit_tag)
);`)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS channel_to_character (
	channel_id TEXT NOT NULL,
	platform_id TEXT NOT NULL,
	character_id TEXT NOT NULL,
	PRIMARY KEY (channel_id, platform_id, character_id)
);`)
	return err
}

func New(ctx context.Context, log *slog.Logger, storagePath string) (*Storage, error) {
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
		{deleteChannelOutfit, "DELETE FROM channel_to_outfit WHERE channel_id = ? AND platform_id = ? AND outfit_tag = lower(?)"},
		{insertChannelCharacter, "INSERT INTO channel_to_character VALUES (?, ?, ?)"},
		{deleteChannelCharacter, "DELETE FROM channel_to_character WHERE channel_id = ? AND platform_id = ? AND character_id = ?"},
		{
			name: selectChannelsByCharacter,
			stmt: `SELECT channel_id FROM channel_to_character WHERE character_id = ?
				   UNION
				   SELECT channel_id FROM channel_to_outfit WHERE outfit_tag = lower(?)`,
		},
		{selectOutfitsForPlatform, "SELECT outfit_tag FROM channel_to_outfit WHERE channel_id = ? AND platform_id = ?"},
		{selectCharactersForPlatform, "SELECT character_id FROM channel_to_character WHERE channel_id = ? AND platform_id = ?"},
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

func (s *Storage) Begin(ctx context.Context) (*Storage, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &Storage{
		log:        s.log,
		db:         s.db,
		statements: s.statements,
		tx:         tx,
	}, nil
}

func (s *Storage) Commit() error {
	if s.tx == nil {
		return ErrTransactionNotStarted
	}
	return s.tx.Commit()
}

func (s *Storage) Rollback() error {
	if s.tx == nil {
		return ErrTransactionNotStarted
	}
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

func (s *Storage) SaveChannelOutfit(ctx context.Context, channelId, platformId, outfitID string) error {
	const op = "storage.sqlite.SaveChannelOutfit"
	err := s.exec(ctx, insertChannelOutfit, channelId, platformId, outfitID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) DeleteChannelOutfit(ctx context.Context, channelId, platformId, outfitID string) error {
	const op = "storage.sqlite.DeleteChannelOutfit"
	err := s.exec(ctx, deleteChannelOutfit, channelId, platformId, outfitID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) SaveChannelCharacter(ctx context.Context, channelId, platformId, characterID string) error {
	const op = "storage.sqlite.SaveChannelCharacter"
	err := s.exec(ctx, insertChannelCharacter, channelId, platformId, characterID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) DeleteChannelCharacter(ctx context.Context, channelId, platformId, characterID string) error {
	const op = "storage.sqlite.DeleteChannelCharacter"
	err := s.exec(ctx, deleteChannelCharacter, channelId, platformId, characterID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
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

func (s *Storage) TrackingOutfitsForPlatform(ctx context.Context, channelId, platformId string) ([]string, error) {
	const op = "storage.sqlite.TrackingOutfitsForPlatform"
	rows, err := query[string](ctx, s, selectOutfitsForPlatform, channelId, platformId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (s *Storage) TrackingCharactersForPlatform(ctx context.Context, channelId, platformId string) ([]string, error) {
	const op = "storage.sqlite.TrackingCharactersForPlatform"
	rows, err := query[string](ctx, s, selectCharactersForPlatform, channelId, platformId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}
