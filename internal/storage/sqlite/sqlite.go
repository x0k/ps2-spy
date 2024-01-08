package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func (s *Storage) prepareAndExec(ctx context.Context, statement string, args ...any) error {
	st, err := s.db.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	_, err = st.ExecContext(ctx, args...)
	if err != nil {
		return err
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
	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) SaveGuildToOutfitRelation(ctx context.Context, guildID, platformId, outfitID string) error {
	const op = "storage.sqlite.SaveGuildToOutfitRelation"
	err := s.prepareAndExec(ctx, "INSERT INTO guild_to_outfit VALUES (?, ?, ?)", guildID, platformId, outfitID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) SaveGuildToCharacterRelation(ctx context.Context, guildID, platformId, characterID string) error {
	const op = "storage.sqlite.SaveGuildToCharacterRelation"
	err := s.prepareAndExec(ctx, "INSERT INTO guild_to_character VALUES (?, ?, ?)", guildID, platformId, characterID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
