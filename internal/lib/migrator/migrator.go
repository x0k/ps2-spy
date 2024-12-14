package migrator

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
)

type Migrator struct {
	log           *slog.Logger
	connectionURI string
	migrationsURI string
	timeout       time.Duration
	attempts      int
}

func New(
	log *slog.Logger,
	connectionURI string,
	migrationsURI string,
) *Migrator {
	return &Migrator{
		log:           log,
		connectionURI: connectionURI,
		migrationsURI: migrationsURI,
		attempts:      3,
		timeout:       1 * time.Second,
	}
}

func (m *Migrator) Migrate(ctx context.Context) error {
	var (
		attempt int
		err     error
		mg      *migrate.Migrate
	)

	for {
		mg, err = migrate.New(m.migrationsURI, m.connectionURI)
		if err == nil {
			break
		}
		attempt++
		m.log.LogAttrs(
			ctx, slog.LevelError, "can't connect",
			slog.Int("attempt", attempt),
			slog.String("error", err.Error()),
		)
		time.Sleep(m.timeout)
		if attempt >= m.attempts {
			return err
		}
	}

	err = mg.Up()
	defer mg.Close()
	if errors.Is(err, migrate.ErrNoChange) {
		m.log.LogAttrs(ctx, slog.LevelInfo, "no change")
		return nil
	}
	if err != nil {
		m.log.LogAttrs(ctx, slog.LevelError, "up error", slog.String("error", err.Error()))
		return err
	}
	m.log.LogAttrs(ctx, slog.LevelInfo, "up success")
	return nil
}
