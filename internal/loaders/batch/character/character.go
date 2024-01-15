package character

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

type CharacterLoader struct {
	wg          sync.WaitGroup
	log         *slog.Logger
	cache       *expirable.LRU[string, ps2.Character]
	loader      loaders.QueriedLoader[[]string, map[string]ps2.Character]
	charId      chan string
	charIdBatch chan []string
	awaitersMu  sync.Mutex
	awaiters    map[string]chan ps2.Character
	batchRate   time.Duration
}

func New(
	log *slog.Logger,
	loader loaders.QueriedLoader[[]string, map[string]ps2.Character],
) *CharacterLoader {
	return &CharacterLoader{
		cache:       expirable.NewLRU[string, ps2.Character](1000, nil, time.Hour*12),
		log:         log.With(slog.String("loaders.batch", "character")),
		loader:      loader,
		charId:      make(chan string, 1000),
		charIdBatch: make(chan []string),
		awaiters:    make(map[string]chan ps2.Character),
	}
}

func (l *CharacterLoader) batcher(ctx context.Context) {
	defer l.wg.Done()
	timer := time.NewTimer(l.batchRate)
	defer timer.Stop()
	var batch []string
	for {
		select {
		case <-ctx.Done():
			return
		case id := <-l.charId:
			batch = append(batch, id)
		case <-timer.C:
			if len(batch) == 0 {
				continue
			}
			l.charIdBatch <- batch
			batch = make([]string, 0, len(batch))
		}
	}
}

func (l *CharacterLoader) releaseAwaiters(batch []string, chars map[string]ps2.Character) {
	l.awaitersMu.Lock()
	defer l.awaitersMu.Unlock()
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

func (l *CharacterLoader) worker(ctx context.Context) {
	defer l.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case batch := <-l.charIdBatch:
			loaded, err := l.loader.Load(ctx, batch)
			if err != nil {
				l.log.Error("failed to load characters", sl.Err(err))
				loaded = make(map[string]ps2.Character)
			}
			l.releaseAwaiters(batch, loaded)
		}
	}
}

func (l *CharacterLoader) load(charId string) chan ps2.Character {
	l.awaitersMu.Lock()
	defer l.awaitersMu.Unlock()
	c := make(chan ps2.Character)
	l.awaiters[charId] = c
	l.charId <- charId
	return c
}

func (l *CharacterLoader) Start(ctx context.Context) {
	l.log.Info("starting loader")
	l.wg.Add(2)
	go l.batcher(ctx)
	go l.worker(ctx)
}

func (l *CharacterLoader) Stop() {
	l.log.Info("stopping loader")
	l.awaitersMu.Lock()
	defer l.awaitersMu.Unlock()
	for _, c := range l.awaiters {
		close(c)
	}
	clear(l.awaiters)
	l.wg.Wait()
}

func (l *CharacterLoader) Load(ctx context.Context, charId string) (ps2.Character, error) {
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
