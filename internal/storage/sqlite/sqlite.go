package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/x0k/ps2-spy/internal/storage"
)

const (
	upsertGuildPlatform = iota
	selectGuildPlatform
	insertGuildOutfit
	insertGuildCharacter
	statementsCount
)

type Storage struct {
	log        *slog.Logger
	db         *sql.DB
	statements [statementsCount]*sql.Stmt
}

func (s *Storage) migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS guild_to_outfit (
	guild_id TEXT NOT NULL,
	platform_id TEXT NOT NULL,
	outfit_id TEXT NOT NULL,
	PRIMARY KEY (guild_id, platform_id, outfit_id)
);`)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS guild_to_character (
	guild_id TEXT NOT NULL,
	platform_id TEXT NOT NULL,
	character_id TEXT NOT NULL,
	PRIMARY KEY (guild_id, platform_id, character_id)
);`)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS guild_to_platform (
	guild_id TEXT PRIMARY KEY NOT NULL,
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
	s := &Storage{db: db, log: log.With(slog.String("component", "sqlite"))}
	if err = s.migrate(ctx); err != nil {
		return nil, fmt.Errorf("%s cannot migrate: %w", op, err)
	}
	rawStatements := [statementsCount]struct {
		name int
		stmt string
	}{
		{upsertGuildPlatform, "INSERT INTO guild_to_platform VALUES (?, ?) ON CONFLICT(guild_id) DO UPDATE SET platform_id = EXCLUDED.platform_id"},
		{selectGuildPlatform, "SELECT platform_id FROM guild_to_platform WHERE guild_id = ?"},
		{insertGuildOutfit, "INSERT INTO guild_to_outfit VALUES (?, ?, ?)"},
		{insertGuildCharacter, "INSERT INTO guild_to_character VALUES (?, ?, ?)"},
	}
	for _, raw := range rawStatements {
		stmt, err := db.Prepare(raw.stmt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		s.statements[raw.name] = stmt
	}
	return s, nil
}

func (s *Storage) Close() error {
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

func (s *Storage) SaveGuildPlatform(ctx context.Context, guildID, platformID string) error {
	const op = "storage.sqlite.SaveGuildPlatform"
	err := s.exec(ctx, upsertGuildPlatform, guildID, platformID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetGuildPlatform(ctx context.Context, guildID string) (string, error) {
	const op = "storage.sqlite.GetGuildPlatform"
	var platformID string
	err := s.queryRow(ctx, &platformID, selectGuildPlatform, guildID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return platformID, nil
}

func (s *Storage) SaveGuildOutfit(ctx context.Context, guildID, platformId, outfitID string) error {
	const op = "storage.sqlite.SaveGuildOutfit"
	err := s.exec(ctx, insertGuildOutfit, guildID, platformId, outfitID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) SaveGuildCharacter(ctx context.Context, guildID, platformId, characterID string) error {
	const op = "storage.sqlite.SaveGuildCharacter"
	err := s.exec(ctx, insertGuildCharacter, guildID, platformId, characterID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
