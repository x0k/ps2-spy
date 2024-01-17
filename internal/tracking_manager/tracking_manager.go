package tracking_manager

import (
	"context"
	"fmt"
	"sync"

	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

var ErrUnknownEvent = fmt.Errorf("unknown event")

type TrackingManager struct {
	charactersFilterMu sync.RWMutex
	charactersFilter   map[string]int
	characterLoader    loaders.KeyedLoader[string, ps2.Character]
	channelsLoader     loaders.KeyedLoader[ps2.Character, []string]
}

func New(
	charLoader loaders.KeyedLoader[string, ps2.Character],
	channelsLoader loaders.KeyedLoader[ps2.Character, []string],
) *TrackingManager {
	return &TrackingManager{
		charactersFilter: make(map[string]int),
		characterLoader:  charLoader,
		channelsLoader:   channelsLoader,
	}
}

func (tm *TrackingManager) isCharacterTracked(charId string) bool {
	tm.charactersFilterMu.RLock()
	defer tm.charactersFilterMu.RUnlock()
	_, ok := tm.charactersFilter[charId]
	return ok
}

func (tm *TrackingManager) ChannelIds(ctx context.Context, event any) ([]string, error) {
	const op = "TrackingManager.ChannelIds"
	switch e := event.(type) {
	case ps2events.PlayerLogin:
		if !tm.isCharacterTracked(e.CharacterID) {
			return nil, nil
		}
		char, err := tm.characterLoader.Load(ctx, e.CharacterID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		return tm.channelsLoader.Load(ctx, char)
	}
	return nil, fmt.Errorf("%s: %w", op, ErrUnknownEvent)
}
