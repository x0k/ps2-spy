package characters_tracker

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
	"github.com/x0k/ps2-spy/internal/lib/retryable"
	"github.com/x0k/ps2-spy/internal/lib/retryable/perform"
	"github.com/x0k/ps2-spy/internal/lib/retryable/while"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/metrics"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/factions"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

var ErrWorldPopulationTrackerNotFound = fmt.Errorf("world population tracker not found")

type player struct {
	characterId ps2.CharacterId
	worldId     ps2.WorldId
}

type CharactersTracker struct {
	log                      *logger.Logger
	platform                 platforms.Platform
	mutex                    sync.RWMutex
	worldPopulationTrackers  map[ps2.WorldId]worldPopulationTracker
	onlineCharactersTracker  onlineCharactersTracker
	activePlayers            *containers.ExpirationQueue[player]
	inactivityCheckInterval  time.Duration
	inactiveTimeout          time.Duration
	retryableCharacterLoader *retryable.WithArg[ps2.CharacterId, ps2.Character]
	publisher                publisher.Abstract[publisher.Event]
	mt                       metrics.Metrics
}

func New(
	log *logger.Logger,
	platform platforms.Platform,
	worldIds []ps2.WorldId,
	characterLoader loaders.KeyedLoader[ps2.CharacterId, ps2.Character],
	publisher publisher.Abstract[publisher.Event],
	mt metrics.Metrics,
) *CharactersTracker {
	trackers := make(map[ps2.WorldId]worldPopulationTracker, len(ps2.WorldNames))
	for _, worldId := range worldIds {
		trackers[worldId] = newWorldPopulationTracker()
	}
	return &CharactersTracker{
		log: log.With(
			slog.String("component", "characters_tracker.CharactersTracker"),
			slog.String("platform", string(platform)),
		),
		platform:                platform,
		worldPopulationTrackers: trackers,
		onlineCharactersTracker: newOnlineCharactersTracker(),
		activePlayers:           containers.NewExpirationQueue[player](),
		inactivityCheckInterval: time.Minute,
		inactiveTimeout:         10 * time.Minute,
		retryableCharacterLoader: retryable.NewWithArg[ps2.CharacterId, ps2.Character](
			characterLoader.Load,
		),
		publisher: publisher,
		mt:        mt,
	}
}

func (p *CharactersTracker) handleInactive(ctx context.Context, now time.Time) []player {
	removedPlayers := make([]player, 0)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	count := p.activePlayers.RemoveExpired(now.Add(-p.inactiveTimeout), func(pl player) {
		if w, ok := p.worldPopulationTrackers[pl.worldId]; ok {
			w.HandleInactive(pl.characterId)
		} else {
			p.log.Warn(ctx, "world not found", slog.String("world_id", string(pl.worldId)))
		}
		if p.onlineCharactersTracker.HandleInactive(pl.characterId) {
			removedPlayers = append(removedPlayers, pl)
		}
	})
	p.mt.SetPlatformQueueSize(
		metrics.ActivePlayersQueueName,
		p.platform,
		p.activePlayers.Len(),
	)
	if count > 0 {
		p.log.Debug(
			ctx,
			"inactive players removed",
			slog.Int("queue_size", p.activePlayers.Len()),
			slog.Int("count", count),
		)
	}
	return removedPlayers
}

func (p *CharactersTracker) publishPlayerLogout(ctx context.Context, t time.Time, pl player) {
	if err := p.publisher.Publish(PlayerLogout{
		Time:        t,
		CharacterId: pl.characterId,
		WorldId:     pl.worldId,
	}); err != nil {
		p.log.Error(ctx, "cannot publish event", sl.Err(err))
	}
}

func (p *CharactersTracker) Start(ctx context.Context) {
	wg := infra.Wg(ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(p.inactivityCheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				removedPlayers := p.handleInactive(ctx, t)
				for _, pl := range removedPlayers {
					p.publishPlayerLogout(ctx, t, pl)
				}
			}
		}
	}()
}

func (p *CharactersTracker) handleLogin(ctx context.Context, char ps2.Character) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.activePlayers.Push(player{char.Id, char.WorldId})
	if w, ok := p.worldPopulationTrackers[char.WorldId]; ok {
		w.HandleLogin(char)
	} else {
		p.log.Warn(ctx, "world not found", slog.String("world_id", string(char.WorldId)))
	}
	return p.onlineCharactersTracker.HandleLogin(char)
}

func (p *CharactersTracker) publishPlayerLogin(ctx context.Context, event ps2events.PlayerLogin) {
	var t time.Time
	// TODO: extract this somewhere
	if timestamp, err := strconv.ParseInt(event.Timestamp, 10, 64); err == nil {
		t = time.Unix(timestamp, 0)
	} else {
		t = time.Now()
	}
	if err := p.publisher.Publish(PlayerLogin{
		Time:        t,
		CharacterId: ps2.CharacterId(event.CharacterID),
		WorldId:     ps2.WorldId(event.WorldID),
	}); err != nil {
		p.log.Error(ctx, "cannot publish event", sl.Err(err))
	}
}

func (p *CharactersTracker) HandleLoginTask(ctx context.Context, wg *sync.WaitGroup, event ps2events.PlayerLogin) {
	defer wg.Done()
	charId := ps2.CharacterId(event.CharacterID)
	char, err := p.retryableCharacterLoader.Run(
		ctx,
		charId,
		while.ErrorIsHere,
		while.RetryCountIsLessThan(3),
		while.ContextIsNotCancelled,
		perform.Log(
			p.log.Logger,
			slog.LevelDebug,
			"[ERROR] failed to get character, retrying",
			slog.String("character_id", string(charId)),
		),
	)
	if err != nil {
		p.log.Error(ctx, "failed to get character", slog.String("character_id", string(charId)), sl.Err(err))
		return
	}
	if p.handleLogin(ctx, char) {
		p.publishPlayerLogin(ctx, event)
	}
}

func (p *CharactersTracker) handleLogout(ctx context.Context, charId ps2.CharacterId, worldId ps2.WorldId) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.activePlayers.Remove(player{charId, worldId})
	if w, ok := p.worldPopulationTrackers[worldId]; ok {
		w.HandleInactive(charId)
	} else {
		p.log.Warn(ctx, "world not found", slog.String("world_id", string(worldId)))
	}
	return p.onlineCharactersTracker.HandleInactive(charId)
}

func (p *CharactersTracker) HandleLogout(ctx context.Context, event ps2events.PlayerLogout) {
	worldId := ps2.WorldId(event.WorldID)
	charId := ps2.CharacterId(event.CharacterID)
	if p.handleLogout(ctx, charId, worldId) {
		var t time.Time
		if timestamp, err := strconv.ParseInt(event.Timestamp, 10, 64); err == nil {
			t = time.Unix(timestamp, 0)
		} else {
			t = time.Now()
		}
		p.publishPlayerLogout(ctx, t, player{charId, worldId})
	}
}

func (p *CharactersTracker) HandleWorldZoneAction(ctx context.Context, worldId, zoneId, charId string) {
	cId := ps2.CharacterId(charId)
	wId := ps2.WorldId(worldId)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	// TODO: Generate login (or something similar) event if char is not logged in yet
	p.activePlayers.Push(player{cId, wId})
	if w, ok := p.worldPopulationTrackers[wId]; ok {
		w.HandleZoneAction(cId, zoneId)
	} else {
		p.log.Warn(ctx, "world not found", slog.String("world_id", worldId))
	}
}

func (p *CharactersTracker) TrackableOnlineEntities(settings meta.SubscriptionSettings) meta.TrackableEntities[map[ps2.OutfitId][]ps2.Character, []ps2.Character] {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.onlineCharactersTracker.TrackableOnlineEntities(settings)
}

func (p *CharactersTracker) WorldsPopulation() ps2.WorldsPopulation {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	total := 0
	worlds := make([]ps2.WorldPopulation, 0, len(p.worldPopulationTrackers))
	for worldId, worldTracker := range p.worldPopulationTrackers {
		worldPopulation := worldTracker.Population()
		other := worldPopulation[factions.None]
		vs := worldPopulation[factions.VS]
		nc := worldPopulation[factions.NC]
		tr := worldPopulation[factions.TR]
		ns := worldPopulation[factions.NSO]
		all := vs + nc + tr + ns + other
		total += all
		worlds = append(worlds, ps2.WorldPopulation{
			Id:   worldId,
			Name: ps2.WorldNames[worldId],
			StatPerFactions: ps2.StatPerFactions{
				All:   all,
				VS:    vs,
				NC:    nc,
				TR:    tr,
				NS:    ns,
				Other: other,
			},
		})
	}
	return ps2.WorldsPopulation{
		Total:  total,
		Worlds: worlds,
	}
}

func (p *CharactersTracker) DetailedWorldPopulation(worldId ps2.WorldId) (ps2.DetailedWorldPopulation, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	tracker, ok := p.worldPopulationTrackers[worldId]
	if !ok {
		return ps2.DetailedWorldPopulation{}, ErrWorldPopulationTrackerNotFound
	}
	zonesPopulation := tracker.ZonesPopulation()
	total := 0
	zones := make([]ps2.ZonePopulation, 0, len(zonesPopulation))
	for zoneId, zonePopulation := range zonesPopulation {
		other := zonePopulation[factions.None]
		vs := zonePopulation[factions.VS]
		nc := zonePopulation[factions.NC]
		tr := zonePopulation[factions.TR]
		ns := zonePopulation[factions.NSO]
		all := vs + nc + tr + ns + other
		total += all
		zones = append(zones, ps2.ZonePopulation{
			Id:   zoneId,
			Name: ps2.ZoneNames[zoneId],
			// TODO: Track this
			IsOpen: all > 0,
			StatPerFactions: ps2.StatPerFactions{
				All:   all,
				VS:    vs,
				NC:    nc,
				TR:    tr,
				NS:    ns,
				Other: other,
			},
		})
	}
	return ps2.DetailedWorldPopulation{
		Id:    worldId,
		Name:  ps2.WorldNameById(worldId),
		Total: total,
		Zones: zones,
	}, nil
}
