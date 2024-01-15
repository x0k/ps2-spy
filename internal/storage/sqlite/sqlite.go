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
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/storage"
)

const (
	upsertChatPlatform = iota
	selectChatPlatform
	insertChatOutfit
	insertChatCharacter
	selectChatIdsByCharacter
	statementsCount
)

type Storage struct {
	log        *slog.Logger
	db         *sql.DB
	statements [statementsCount]*sql.Stmt
}

func (s *Storage) migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS chat_to_outfit (
	chat_id TEXT NOT NULL,
	platform_id TEXT NOT NULL,
	outfit_tag TEXT NOT NULL,
	PRIMARY KEY (chat_id, platform_id, outfit_tag)
);`)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS chat_to_character (
	chat_id TEXT NOT NULL,
	platform_id TEXT NOT NULL,
	character_id TEXT NOT NULL,
	PRIMARY KEY (chat_id, platform_id, character_id)
);`)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS chat_to_platform (
	chat_id TEXT PRIMARY KEY NOT NULL,
	platform_id TEXT NOT NULL
);`)
	if err != nil {
		return err
	}
	return nil
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
		{upsertChatPlatform, "INSERT INTO chat_to_platform VALUES (?, ?) ON CONFLICT(chat_id) DO UPDATE SET platform_id = EXCLUDED.platform_id"},
		{selectChatPlatform, "SELECT platform_id FROM chat_to_platform WHERE chat_id = ?"},
		{insertChatOutfit, "INSERT INTO chat_to_outfit VALUES (?, ?, lower(?))"},
		{insertChatCharacter, "INSERT INTO chat_to_character VALUES (?, ?, ?)"},
		{
			name: selectChatIdsByCharacter,
			stmt: `SELECT chat_id FROM chat_to_character WHERE character_id = ?
				   UNION
				   SELECT chat_id FROM chat_to_outfit WHERE outfit_tag = lower(?)`,
		},
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

func (s *Storage) exec(ctx context.Context, statement int, args ...any) error {
	if _, err := s.statements[statement].ExecContext(ctx, args...); err != nil {
		return err
	}
	return nil
}

func (s *Storage) queryRow(ctx context.Context, result any, statement int, args ...any) error {
	if err := s.statements[statement].QueryRowContext(ctx, args...).Scan(result); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrNotFound
		}
		return err
	}
	return nil
}

func query[T any](ctx context.Context, s *Storage, statement int, args ...any) ([]T, error) {
	rows, err := s.statements[statement].QueryContext(ctx, args...)
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

func (s *Storage) SaveChatPlatform(ctx context.Context, chatId, platformID string) error {
	const op = "storage.sqlite.SaveChatPlatform"
	err := s.exec(ctx, upsertChatPlatform, chatId, platformID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetChatPlatform(ctx context.Context, chatId string) (string, error) {
	const op = "storage.sqlite.GetChatPlatform"
	var platformID string
	err := s.queryRow(ctx, &platformID, selectChatPlatform, chatId)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return platformID, nil
}

func (s *Storage) SaveChatOutfit(ctx context.Context, chatId, platformId, outfitID string) error {
	const op = "storage.sqlite.SaveChatOutfit"
	err := s.exec(ctx, insertChatOutfit, chatId, platformId, outfitID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) SaveChatCharacter(ctx context.Context, chatId, platformId, characterID string) error {
	const op = "storage.sqlite.SaveChatCharacter"
	err := s.exec(ctx, insertChatCharacter, chatId, platformId, characterID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) TrackingChannelIdsForCharacter(ctx context.Context, character ps2.Character) ([]string, error) {
	const op = "storage.sqlite.TrackingChannelIdsForCharacter"
	rows, err := query[string](ctx, s, selectChatIdsByCharacter, character.Id, character.OutfitTag)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}
