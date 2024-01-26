package character_loader

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type BatchLoader struct {
	cache      *expirable.LRU[ps2.CharacterId, ps2.Character]
	loader     loaders.QueriedLoader[[]ps2.CharacterId, map[ps2.CharacterId]ps2.Character]
	awaitersMu sync.Mutex
	awaiters   map[ps2.CharacterId][]chan ps2.Character
	batchRate  time.Duration
}

func NewBatch(
	loader loaders.QueriedLoader[[]ps2.CharacterId, map[ps2.CharacterId]ps2.Character],
	batchRate time.Duration,
) *BatchLoader {
	return &BatchLoader{
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

func (l *BatchLoader) releaseAwaiters(log *slog.Logger, batch []ps2.CharacterId, chars map[ps2.CharacterId]ps2.Character) {
	l.awaitersMu.Lock()
	defer l.awaitersMu.Unlock()
	if len(batch) != len(chars) {
		log.Warn(
			"not all characters were loaded",
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
			log.Warn("awaiter not found", slog.String("character_id", string(id)))
		}
	}
}

func (l *BatchLoader) processBatchTask(ctx context.Context, wg *sync.WaitGroup) {
	const op = "loaders.character_loader.batch.processBatchTask"
	log := infra.OpLogger(ctx, op)
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
			log.Debug("execute batch", slog.Int("batch_size", len(batch)))
			loaded, err := l.loader.Load(ctx, batch)
			if err != nil {
				log.Error("failed to load characters", sl.Err(err))
				loaded = make(map[ps2.CharacterId]ps2.Character)
			}
			l.releaseAwaiters(log, batch, loaded)
		}
	}
}

func (l *BatchLoader) cleanupTask(ctx context.Context, wg *sync.WaitGroup) {
	const op = "loaders.character_loader.batch.cleanupTask"
	log := infra.OpLogger(ctx, op)
	defer wg.Done()
	<-ctx.Done()
	log.Info("cleaning up")
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
	infra.OpLogger(ctx, op).Info("starting loader")
	wg.Add(2)
	go l.processBatchTask(ctx, wg)
	go l.cleanupTask(ctx, wg)
}

func (l *BatchLoader) load(log *slog.Logger, charId ps2.CharacterId) chan ps2.Character {
	l.awaitersMu.Lock()
	defer l.awaitersMu.Unlock()
	log.Debug("awaiting for", slog.String("character_id", string(charId)))
	c := make(chan ps2.Character)
	l.awaiters[charId] = append(l.awaiters[charId], c)
	return c
}

func (l *BatchLoader) Load(ctx context.Context, charId ps2.CharacterId) (ps2.Character, error) {
	const op = "loaders.character_loader.batch.Load"
	log := infra.OpLogger(ctx, op)
	cached, ok := l.cache.Get(charId)
	if ok {
		return cached, nil
	}
	select {
	case <-ctx.Done():
		return ps2.Character{}, ctx.Err()
	case char := <-l.load(log, charId):
		if char.Id == "" {
			return char, loaders.ErrNotFound
		}
		l.cache.Add(charId, char)
		return char, nil
	}
}
