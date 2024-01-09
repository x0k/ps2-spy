package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/x0k/ps2-spy/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func (s *Storage) prepareAndExec(ctx context.Context, statement string, args ...any) error {
	st, err := s.db.PrepareContext(ctx, statement)
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	_, err = st.ExecContext(ctx, args...)
	if err != nil {
		return fmt.Errorf("error executing statement: %w", err)
	}
	return nil
}

func preparedAndQueryRow[R any](ctx context.Context, db *sql.DB, result R, statement string, args ...any) error {
	st, err := db.PrepareContext(ctx, statement)
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	err = st.QueryRowContext(ctx, args...).Scan(result)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrNotFound
		}
		return fmt.Errorf("error executing statement: %w", err)
	}
	return nil
}

func New(ctx context.Context, storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Migrate(ctx context.Context) error {
	const op = "storage.sqlite.Migrate"
	err := s.prepareAndExec(ctx, `
CREATE TABLE IF NOT EXISTS guild_to_outfit (
	guild_id TEXT NOT NULL,
	platform_id TEXT NOT NULL,
	outfit_id TEXT NOT NULL,
	PRIMARY KEY (guild_id, platform_id, outfit_id)
);`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = s.prepareAndExec(ctx, `
CREATE TABLE IF NOT EXISTS guild_to_character (
	guild_id TEXT NOT NULL,
	platform_id TEXT NOT NULL,
	character_id TEXT NOT NULL,
	PRIMARY KEY (guild_id, platform_id, character_id)
);`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = s.prepareAndExec(ctx, `
CREATE TABLE IF NOT EXISTS guild_to_platform (
	guild_id TEXT PRIMARY KEY NOT NULL,
	platform_id TEXT NOT NULL
);`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) SaveGuildPlatform(ctx context.Context, guildID, platformID string) error {
	const op = "storage.sqlite.SaveGuildPlatform"
	err := s.prepareAndExec(ctx, "INSERT INTO guild_to_platform VALUES (?, ?) ON CONFLICT(guild_id) DO UPDATE SET platform_id = EXCLUDED.platform_id", guildID, platformID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetGuildPlatform(ctx context.Context, guildID string) (string, error) {
	const op = "storage.sqlite.GetGuildPlatform"
	var platformID string
	err := preparedAndQueryRow(
		ctx,
		s.db,
		&platformID,
		"SELECT platform_id FROM guild_to_platform WHERE guild_id = ?",
		guildID,
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return platformID, nil
}

func (s *Storage) SaveGuildOutfit(ctx context.Context, guildID, platformId, outfitID string) error {
	const op = "storage.sqlite.SaveGuildOutfit"
	err := s.prepareAndExec(ctx, "INSERT INTO guild_to_outfit VALUES (?, ?, ?)", guildID, platformId, outfitID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) SaveGuildCharacter(ctx context.Context, guildID, platformId, characterID string) error {
	const op = "storage.sqlite.SaveGuildCharacter"
	err := s.prepareAndExec(ctx, "INSERT INTO guild_to_character VALUES (?, ?, ?)", guildID, platformId, characterID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
