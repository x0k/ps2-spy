package ps2events

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
)

var ErrEventNameNotFound = fmt.Errorf("event name not found")
var ErrUnknownEventName = fmt.Errorf("unknown event name")

type Publisher struct {
	log        *slog.Logger
	handlersMu sync.RWMutex
	handlers   map[string][]eventHandler
	buffers    map[string]any
}

func NewPublisher(log *slog.Logger) *Publisher {
	return &Publisher{
		log: log.With(
			slog.String("component", "census2.streaming.events.Publisher"),
		),
		handlers: map[string][]eventHandler{},
		buffers: map[string]any{
			AchievementEarnedEventName:     &AchievementEarned{},
			BattleRankUpEventName:          &BattleRankUp{},
			DeathEventName:                 &Death{},
			GainExperienceEventName:        &GainExperience{},
			ItemAddedEventName:             &ItemAdded{},
			PlayerFacilityCaptureEventName: &PlayerFacilityCapture{},
			PlayerFacilityDefendEventName:  &PlayerFacilityDefend{},
			PlayerLoginEventName:           &PlayerLogin{},
			PlayerLogoutEventName:          &PlayerLogout{},
			SkillAddedEventName:            &SkillAdded{},
			VehicleDestroyEventName:        &VehicleDestroy{},
			ContinentLockEventName:         &ContinentLock{},
			FacilityControlEventName:       &FacilityControl{},
			MetagameEventEventName:         &MetagameEvent{},
		},
	}
}

func (p *Publisher) removeHandler(eventType string, h eventHandler) {
	p.handlersMu.Lock()
	defer p.handlersMu.Unlock()
	for i, v := range p.handlers[eventType] {
		if v == h {
			p.handlers[eventType] = append(p.handlers[eventType][:i], p.handlers[eventType][i+1:]...)
			return
		}
	}
}

func (p *Publisher) addHandler(h eventHandler) func() {
	p.handlersMu.Lock()
	defer p.handlersMu.Unlock()
	p.handlers[h.Type()] = append(p.handlers[h.Type()], h)
	return func() {
		p.removeHandler(h.Type(), h)
	}
}

func (p *Publisher) AddHandler(h any) (func(), error) {
	handler := handlerForInterface(h)
	if handler == nil {
		return nil, ErrUnknownEventName
	}
	return p.addHandler(handler), nil
}

func (p *Publisher) publish(eventType string, msg any) {
	p.handlersMu.RLock()
	defer p.handlersMu.RUnlock()
	for _, h := range p.handlers[eventType] {
		h.Handle(msg)
	}
}

func (p *Publisher) Publish(event map[string]any) {
	var err error
	defer func() {
		if err != nil {
			p.log.Warn("failed to publish event", slog.Any("event", event), sl.Err(err))
		}
	}()
	name, ok := event[core.EventNameField].(string)
	if !ok {
		err = ErrEventNameNotFound
		return
	}
	if buff, ok := p.buffers[name]; ok {
		err = mapstructure.Decode(event, buff)
		if err != nil {
			return
		}
		p.publish(name, buff)
		return
	}
	err = fmt.Errorf("%s: %w", name, ErrUnknownEventName)
}
