package sql_storage

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/db"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
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

func (s *Storage) Queries() *db.Queries {
	return s.queries
}

func (s *Storage) Transaction(ctx context.Context, run func(s storage.Storage) error) error {
	return s.Begin(ctx, 10, func(tx *Storage) error {
		return run(tx)
	})
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
		log:         s.log,
		db:          s.db,
		queries:     s.queries.WithTx(tx),
		publisher:   bufferedPublisher,
		storagePath: s.storagePath,
	}
	err = run(tmp)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	bufferedPublisher.Flush()
	return nil
}

func (s *Storage) Channel(
	ctx context.Context,
	channelId discord.ChannelId,
) (discord.Channel, error) {
	c, err := s.queries.GetChannel(ctx, string(channelId))
	if err == nil {
		return storage.ChannelFromDTO(ctx, s.log.Logger, c), nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return discord.NewDefaultChannel(channelId), nil
	}
	return discord.Channel{}, err
}

func (s *Storage) SaveChannelLanguage(
	ctx context.Context,
	channelId discord.ChannelId,
	locale language.Tag,
) error {
	err := s.queries.UpsertChannelLanguage(ctx, db.UpsertChannelLanguageParams{
		ChannelID: string(channelId),
		Locale:    locale.String(),
	})
	return s.publish(err, storage.ChannelLanguageSaved{
		ChannelId: channelId,
		Language:  locale,
	})
}

func (s *Storage) SaveChannelCharacterNotifications(
	ctx context.Context,
	channelId discord.ChannelId,
	enabled bool,
) error {
	return s.queries.UpsertChannelCharacterNotifications(ctx, db.UpsertChannelCharacterNotificationsParams{
		ChannelID:              string(channelId),
		CharacterNotifications: enabled,
	})
}

func (s *Storage) SaveChannelOutfitNotifications(
	ctx context.Context,
	channelId discord.ChannelId,
	enabled bool,
) error {
	return s.queries.UpsertChannelOutfitNotifications(ctx, db.UpsertChannelOutfitNotificationsParams{
		ChannelID:           string(channelId),
		OutfitNotifications: enabled,
	})
}

func (s *Storage) SaveChannelTitleUpdates(
	ctx context.Context,
	channelId discord.ChannelId,
	enabled bool,
) error {
	err := s.queries.UpsertChannelTitleUpdates(ctx, db.UpsertChannelTitleUpdatesParams{
		ChannelID:    string(channelId),
		TitleUpdates: enabled,
	})
	return s.publish(err, storage.ChannelTitleUpdatesSaved{
		ChannelId: channelId,
		Enabled:   enabled,
	})
}

func (s *Storage) SaveChannelDefaultTimezone(
	ctx context.Context,
	channelId discord.ChannelId,
	loc *time.Location,
) error {
	return s.queries.UpsertChannelDefaultTimezone(ctx, db.UpsertChannelDefaultTimezoneParams{
		ChannelID:       string(channelId),
		DefaultTimezone: loc.String(),
	})
}

func (s *Storage) publish(err error, event storage.Event) error {
	if errors.Is(err, sql.ErrNoRows) {
		return shared.ErrNotFound
	}
	if err != nil {
		return err
	}
	s.publisher.Publish(event)
	return nil
}
