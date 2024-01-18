package character_loader

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type BatchLoader struct {
	log         *slog.Logger
	cache       *expirable.LRU[string, ps2.Character]
	loader      loaders.QueriedLoader[[]string, map[string]ps2.Character]
	charId      chan string
	charIdBatch chan []string
	awaitersMu  sync.Mutex
	awaiters    map[string]chan ps2.Character
	batchRate   time.Duration
}

func NewBatch(
	log *slog.Logger,
	loader loaders.QueriedLoader[[]string, map[string]ps2.Character],
	batchRate time.Duration,
) *BatchLoader {
	return &BatchLoader{
		log:         log.With(slog.String("loaders.batch", "character")),
		cache:       expirable.NewLRU[string, ps2.Character](1000, nil, time.Hour*12),
		loader:      loader,
		charId:      make(chan string, 1000),
		charIdBatch: make(chan []string),
		awaiters:    make(map[string]chan ps2.Character),
		batchRate:   batchRate,
	}
}

func (l *BatchLoader) batcher(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(l.batchRate)
	defer ticker.Stop()
	var batch []string
	for {
		select {
		case <-ctx.Done():
			return
		case id := <-l.charId:
			batch = append(batch, id)
		case <-ticker.C:
			l.log.Debug("flushing batch", slog.Int("characters", len(batch)))
			if len(batch) == 0 {
				continue
			}
			l.charIdBatch <- batch
			batch = make([]string, 0, len(batch))
		}
	}
}

func (l *BatchLoader) releaseAwaiters(batch []string, chars map[string]ps2.Character) {
	l.awaitersMu.Lock()
	defer l.awaitersMu.Unlock()
	l.log.Debug("releasing", slog.Int("characters", len(batch)))
	for _, id := range batch {
		if c, ok := l.awaiters[id]; ok {
			if char, ok := chars[id]; ok {
				c <- char
			}
			close(c)
			delete(l.awaiters, id)
		}
	}
}

func (l *BatchLoader) worker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case batch := <-l.charIdBatch:
			l.log.Debug("execute batch", slog.Int("characters", len(batch)))
			loaded, err := l.loader.Load(ctx, batch)
			if err != nil {
				l.log.Error("failed to load characters", sl.Err(err))
				loaded = make(map[string]ps2.Character)
			}
			l.releaseAwaiters(batch, loaded)
		}
	}
}

func (l *BatchLoader) cleaner(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	<-ctx.Done()
	l.log.Info("cleaning up")
	l.awaitersMu.Lock()
	defer l.awaitersMu.Unlock()
	for _, c := range l.awaiters {
		close(c)
	}
	clear(l.awaiters)
}

func (l *BatchLoader) Start(ctx context.Context, wg *sync.WaitGroup) {
	l.log.Info("starting loader")
	wg.Add(3)
	go l.batcher(ctx, wg)
	go l.worker(ctx, wg)
	go l.cleaner(ctx, wg)
}

func (l *BatchLoader) load(charId string) chan ps2.Character {
	l.awaitersMu.Lock()
	defer l.awaitersMu.Unlock()
	l.log.Debug("awaiting for", slog.String("character_id", charId))
	c := make(chan ps2.Character)
	l.awaiters[charId] = c
	l.charId <- charId
	return c
}

func (l *BatchLoader) Load(ctx context.Context, charId string) (ps2.Character, error) {
	l.log.Debug("loading", slog.String("character_id", charId))
	cached, ok := l.cache.Get(charId)
	if ok {
		return cached, nil
	}
	char := <-l.load(charId)
	if char.Id == "" {
		return char, loaders.ErrNotFound
	}
	l.cache.Add(charId, char)
	return char, nil
}
