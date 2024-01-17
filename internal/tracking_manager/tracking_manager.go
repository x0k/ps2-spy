package tracking_manager

import (
	"context"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/bits-and-blooms/bloom/v3"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

var ErrUnknownEvent = fmt.Errorf("unknown event")

type TrackingManager struct {
	charactersFilter *bloom.BloomFilter
	characterLoader  loaders.KeyedLoader[string, ps2.Character]
	channelsLoader   loaders.KeyedLoader[ps2.Character, []string]
}

func New(
	charLoader loaders.KeyedLoader[string, ps2.Character],
	channelsLoader loaders.KeyedLoader[ps2.Character, []string],
) *TrackingManager {
	return &TrackingManager{
		charactersFilter: bloom.New(10000, 5),
		characterLoader:  charLoader,
		channelsLoader:   channelsLoader,
	}
}

func strIdToByte(id string) ([]byte, error) {
	number, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, err
	}
	byteSlice := make([]byte, 8)
	binary.BigEndian.PutUint64(byteSlice, number)
	return byteSlice, nil
}

func (tm *TrackingManager) ChannelIds(ctx context.Context, event any) ([]string, error) {
	const op = "TrackingManager.ChannelIds"
	switch e := event.(type) {
	case ps2events.PlayerLogin:
		charId, err := strIdToByte(e.CharacterID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		if !tm.charactersFilter.Test(charId) {
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
