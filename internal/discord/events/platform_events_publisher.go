package discord_events

import (
	"context"
	"log/slog"
	"sync"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

type ChannelsForCharacterLoader = loader.Keyed[ps2.CharacterId, []discord.Channel]
type ChannelsForOutfitLoader = loader.Keyed[ps2.OutfitId, []discord.Channel]

type PlatformEventsPublisher struct {
	publisher                  pubsub.Publisher[Event]
	log                        *logger.Logger
	wg                         sync.WaitGroup
	channelsForCharacterLoader ChannelsForCharacterLoader
	channelsForOutfitLoader    ChannelsForOutfitLoader
}

func NewPlatformEventsPublisher(
	log *logger.Logger,
	publisher pubsub.Publisher[Event],
	channelsForCharacterLoader ChannelsForCharacterLoader,
	channelsForOutfitLoader ChannelsForOutfitLoader,
) *PlatformEventsPublisher {
	return &PlatformEventsPublisher{
		log:                        log,
		publisher:                  publisher,
		channelsForCharacterLoader: channelsForCharacterLoader,
		channelsForOutfitLoader:    channelsForOutfitLoader,
	}
}

func (p *PlatformEventsPublisher) Start(ctx context.Context) {
	<-ctx.Done()
	p.wg.Wait()
}

func (p *PlatformEventsPublisher) PublishPlayerLogin(ctx context.Context, e ps2.PlayerLogin) {
	p.wg.Add(1)
	go publishCharacterEventTask(ctx, p, e.Character.Id, e)
}

func (p *PlatformEventsPublisher) PublishPlayerFakeLogin(ctx context.Context, e ps2.PlayerFakeLogin) {
	p.wg.Add(1)
	go publishCharacterEventTask(ctx, p, e.Character.Id, e)
}

func (p *PlatformEventsPublisher) PublishPlayerLogout(ctx context.Context, e ps2.PlayerLogout) {
	p.wg.Add(1)
	go publishCharacterEventTask(ctx, p, e.CharacterId, e)
}

func (p *PlatformEventsPublisher) PublishFacilityControl(ctx context.Context, e worlds_tracker.FacilityControl) {
	p.wg.Add(1)
	go publishOutfitEventTask(ctx, p, ps2.OutfitId(e.OutfitID), e)
}

func (p *PlatformEventsPublisher) PublishFacilityLoss(ctx context.Context, e worlds_tracker.FacilityLoss) {
	p.wg.Add(1)
	go publishOutfitEventTask(ctx, p, e.OldOutfitId, e)
}

func (p *PlatformEventsPublisher) PublishOutfitMembersUpdate(ctx context.Context, e ps2.OutfitMembersUpdate) {
	p.wg.Add(1)
	go publishOutfitEventTask(ctx, p, e.OutfitId, e)
}

func publishCharacterEventTask[T pubsub.EventType, E pubsub.Event[T]](
	ctx context.Context,
	p *PlatformEventsPublisher,
	characterId ps2.CharacterId,
	event E,
) {
	defer p.wg.Done()
	channels, err := p.channelsForCharacterLoader(ctx, characterId)
	if err != nil {
		p.log.Error(ctx, "cannot get channels for character", slog.String("character_id", string(characterId)), sl.Err(err))
		return
	}
	if len(channels) == 0 {
		return
	}
	p.publisher.Publish(channelsEvent[T, E]{
		Event:    event,
		Channels: channels,
	})
}

func publishOutfitEventTask[T pubsub.EventType, E pubsub.Event[T]](
	ctx context.Context,
	p *PlatformEventsPublisher,
	outfitId ps2.OutfitId,
	event E,
) {
	defer p.wg.Done()
	channels, err := p.channelsForOutfitLoader(ctx, outfitId)
	if err != nil {
		p.log.Error(ctx, "cannot get channels for outfit", slog.String("outfit_id", string(outfitId)), sl.Err(err))
		return
	}
	if len(channels) == 0 {
		return
	}
	p.publisher.Publish(channelsEvent[T, E]{
		Event:    event,
		Channels: channels,
	})
}
