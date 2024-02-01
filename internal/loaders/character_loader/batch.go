package character_loader

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type BatchLoader struct {
	log        *logger.Logger
	cache      *expirable.LRU[ps2.CharacterId, ps2.Character]
	loader     loaders.QueriedLoader[[]ps2.CharacterId, map[ps2.CharacterId]ps2.Character]
	awaitersMu sync.Mutex
	awaiters   map[ps2.CharacterId][]chan ps2.Character
	batchRate  time.Duration
}

func NewBatch(
	log *logger.Logger,
	loader loaders.QueriedLoader[[]ps2.CharacterId, map[ps2.CharacterId]ps2.Character],
	batchRate time.Duration,
) *BatchLoader {
	return &BatchLoader{
		log:       log.With(slog.String("component", "loaders.character_loader.BatchLoader")),
		cache:     expirable.NewLRU[ps2.CharacterId, ps2.Character](0, nil, time.Hour*24),
		loader:    loader,
		awaiters:  make(map[ps2.CharacterId][]chan ps2.Character, 100),
		batchRate: batchRate,
	}
}

func (l *BatchLoader) batch() []ps2.CharacterId {
	l.awaitersMu.Lock()
	defer l.awaitersMu.Unlock()
	batch := make([]ps2.CharacterId, 0, len(l.awaiters))
	for id := range l.awaiters {
		batch = append(batch, id)
	}
	return batch
}

func (l *BatchLoader) releaseAwaiters(ctx context.Context, batch []ps2.CharacterId, chars map[ps2.CharacterId]ps2.Character) {
	l.awaitersMu.Lock()
	defer l.awaitersMu.Unlock()
	if len(batch) != len(chars) {
		l.log.Debug(
			ctx,
			"[WARN] not all characters were loaded",
			slog.Int("batch_size", len(batch)),
			slog.Int("loaded_size", len(chars)),
		)
	}
	for _, id := range batch {
		if channels, ok := l.awaiters[id]; ok {
			char, ok := chars[id]
			for _, c := range channels {
				if ok {
					c <- char
				}
				close(c)
			}
			delete(l.awaiters, id)
		} else {
			l.log.Warn(ctx, "awaiter not found", slog.String("character_id", string(id)))
		}
	}
}

func (l *BatchLoader) processBatchTask(ctx context.Context, wg *sync.WaitGroup) {
	const op = "loaders.character_loader.batch.processBatchTask"
	defer wg.Done()
	ticker := time.NewTicker(l.batchRate)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			batch := l.batch()
			if len(batch) == 0 {
				continue
			}
			l.log.Debug(ctx, "execute batch", slog.Int("batch_size", len(batch)))
			loaded, err := l.loader.Load(ctx, batch)
			if err != nil {
				l.log.Error(ctx, "failed to load characters", sl.Err(err))
				loaded = make(map[ps2.CharacterId]ps2.Character)
			}
			l.releaseAwaiters(ctx, batch, loaded)
		}
	}
}

func (l *BatchLoader) cleanupTask(ctx context.Context, wg *sync.WaitGroup) {
	const op = "loaders.character_loader.batch.cleanupTask"
	defer wg.Done()
	<-ctx.Done()
	l.log.Info(ctx, "cleaning up")
	l.awaitersMu.Lock()
	defer l.awaitersMu.Unlock()
	for _, channels := range l.awaiters {
		for _, c := range channels {
			close(c)
		}
	}
	clear(l.awaiters)
}

func (l *BatchLoader) Start(ctx context.Context, wg *sync.WaitGroup) {
	const op = "loaders.character_loader.batch.Start"
	l.log.Info(ctx, "starting loader")
	wg.Add(2)
	go l.processBatchTask(ctx, wg)
	go l.cleanupTask(ctx, wg)
}

func (l *BatchLoader) load(charId ps2.CharacterId) chan ps2.Character {
	l.awaitersMu.Lock()
	defer l.awaitersMu.Unlock()
	c := make(chan ps2.Character)
	l.awaiters[charId] = append(l.awaiters[charId], c)
	return c
}

func (l *BatchLoader) Load(ctx context.Context, charId ps2.CharacterId) (ps2.Character, error) {
	const op = "loaders.character_loader.batch.Load"
	cached, ok := l.cache.Get(charId)
	if ok {
		return cached, nil
	}
	select {
	case <-ctx.Done():
		return ps2.Character{}, ctx.Err()
	case char := <-l.load(charId):
		if char.Id == "" {
			return char, loaders.ErrNotFound
		}
		l.cache.Add(charId, char)
		return char, nil
	}
}
