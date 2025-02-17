package characters_tracker

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/metrics"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_factions "github.com/x0k/ps2-spy/internal/ps2/factions"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

var ErrWorldPopulationTrackerNotFound = fmt.Errorf("world population tracker not found")

type player struct {
	characterId ps2.CharacterId
	worldId     ps2.WorldId
}

type CharacterLoader = loader.Keyed[ps2.CharacterId, ps2.Character]

type CharactersTracker struct {
	wg                      sync.WaitGroup
	log                     *logger.Logger
	platform                ps2_platforms.Platform
	mutex                   sync.RWMutex
	worldPopulationTrackers map[ps2.WorldId]worldPopulationTracker
	onlineCharactersTracker onlineCharactersTracker
	activePlayers           *containers.ExpirationQueue[player]
	inactivityCheckInterval time.Duration
	inactiveTimeout         time.Duration
	characterLoader         CharacterLoader
	publisher               pubsub.Publisher[Event]
	mt                      *metrics.Metrics
}

func New(
	log *logger.Logger,
	platform ps2_platforms.Platform,
	worldIds []ps2.WorldId,
	characterLoader loader.Keyed[ps2.CharacterId, ps2.Character],
	publisher pubsub.Publisher[Event],
	mt *metrics.Metrics,
) *CharactersTracker {
	trackers := make(map[ps2.WorldId]worldPopulationTracker, len(ps2.WorldNames))
	for _, worldId := range worldIds {
		trackers[worldId] = newWorldPopulationTracker()
	}
	return &CharactersTracker{
		log:                     log,
		platform:                platform,
		worldPopulationTrackers: trackers,
		onlineCharactersTracker: newOnlineCharactersTracker(),
		activePlayers:           containers.NewExpirationQueue[player](),
		inactivityCheckInterval: time.Minute,
		inactiveTimeout:         10 * time.Minute,
		characterLoader:         characterLoader,
		publisher:               publisher,
		mt:                      mt,
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
	metrics.SetPlatformQueueSize(
		p.mt,
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

func (p *CharactersTracker) Start(ctx context.Context) {
	ticker := time.NewTicker(p.inactivityCheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			p.wg.Wait()
			return
		case t := <-ticker.C:
			removedPlayers := p.handleInactive(ctx, t)
			for _, pl := range removedPlayers {
				p.publishPlayerLogout(t, pl)
			}
		}
	}
}

func (p *CharactersTracker) HandleLogin(ctx context.Context, event events.PlayerLogin) {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		charId := ps2.CharacterId(event.CharacterID)
		if char, ok := p.handleLogin(ctx, charId); ok {
			p.publishPlayerLogin(event, char)
		}
	}()
}

func (p *CharactersTracker) HandleLogout(ctx context.Context, event events.PlayerLogout) {
	worldId := ps2.WorldId(event.WorldID)
	charId := ps2.CharacterId(event.CharacterID)
	if p.handleLogout(ctx, charId, worldId) {
		var t time.Time
		if timestamp, err := strconv.ParseInt(event.Timestamp, 10, 64); err == nil {
			t = time.Unix(timestamp, 0)
		} else {
			t = time.Now()
		}
		p.publishPlayerLogout(t, player{charId, worldId})
	}
}

func (p *CharactersTracker) HandleWorldZoneAction(ctx context.Context, worldId, zoneId, charId string) {
	cId := ps2.CharacterId(charId)
	if cId == ps2.RestrictedAreaCharacterId {
		return
	}
	wId := ps2.WorldId(worldId)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.onlineCharactersTracker.isOnline(cId) {
		p.activePlayers.Push(player{cId, wId})
	} else {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			if char, ok := p.handleLogin(ctx, cId); ok {
				p.publishPlayerFakeLogin(char)
			}
		}()
	}
	if w, ok := p.worldPopulationTrackers[wId]; ok {
		w.HandleZoneAction(cId, zoneId)
	} else {
		p.log.Warn(ctx, "world not found", slog.String("world_id", worldId))
	}
}

func (p *CharactersTracker) OutfitMembersOnline(
	outfitIds []ps2.OutfitId,
) map[ps2.OutfitId]map[ps2.CharacterId]ps2.Character {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.onlineCharactersTracker.OutfitMembersOnline(outfitIds)
}

func (p *CharactersTracker) CharactersOnline(
	characterIds []ps2.CharacterId,
) map[ps2.CharacterId]ps2.Character {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.onlineCharactersTracker.CharactersOnline(characterIds)
}

func (p *CharactersTracker) WorldsPopulation() ps2.WorldsPopulation {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	total := 0
	worlds := make([]ps2.WorldPopulation, 0, len(p.worldPopulationTrackers))
	for worldId, worldTracker := range p.worldPopulationTrackers {
		worldPopulation := worldTracker.Population()
		other := worldPopulation[ps2_factions.None]
		vs := worldPopulation[ps2_factions.VS]
		nc := worldPopulation[ps2_factions.NC]
		tr := worldPopulation[ps2_factions.TR]
		ns := worldPopulation[ps2_factions.NSO]
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
		other := zonePopulation[ps2_factions.None]
		vs := zonePopulation[ps2_factions.VS]
		nc := zonePopulation[ps2_factions.NC]
		tr := zonePopulation[ps2_factions.TR]
		ns := zonePopulation[ps2_factions.NSO]
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

func (p *CharactersTracker) handleLogin(ctx context.Context, charId ps2.CharacterId) (ps2.Character, bool) {
	char, err := p.characterLoader(ctx, charId)
	if err != nil {
		p.log.Debug(
			ctx,
			"[ERROR] failed to get character",
			slog.String("character_id", string(charId)),
			sl.Err(err),
		)
		return char, false
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.activePlayers.Push(player{char.Id, char.WorldId})
	if w, ok := p.worldPopulationTrackers[char.WorldId]; ok {
		w.HandleLogin(char)
	} else {
		p.log.Warn(ctx, "world not found", slog.String("world_id", string(char.WorldId)))
	}
	return char, p.onlineCharactersTracker.HandleLogin(char)
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

func (p *CharactersTracker) publishPlayerLogin(
	event events.PlayerLogin,
	char ps2.Character,
) {
	var t time.Time
	// TODO: move this somewhere
	if timestamp, err := strconv.ParseInt(event.Timestamp, 10, 64); err == nil {
		t = time.Unix(timestamp, 0)
	} else {
		t = time.Now()
	}
	p.publisher.Publish(PlayerLogin{
		Time:      t,
		Character: char,
	})
}

func (p *CharactersTracker) publishPlayerFakeLogin(char ps2.Character) {
	now := time.Now()
	p.publisher.Publish(PlayerFakeLogin{
		Time:      now,
		Character: char,
	})
}

func (p *CharactersTracker) publishPlayerLogout(t time.Time, pl player) {
	p.publisher.Publish(PlayerLogout{
		Time:        t,
		CharacterId: pl.characterId,
		WorldId:     pl.worldId,
	})
}
