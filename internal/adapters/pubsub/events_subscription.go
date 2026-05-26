package pubsub_adapters

import (
	"context"

	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

type EventsSubscriptionChannels struct {
	PlayerLogin           <-chan events.PlayerLogin
	PlayerLogout          <-chan events.PlayerLogout
	AchievementEarned     <-chan events.AchievementEarned
	BattleRankUp          <-chan events.BattleRankUp
	Death                 <-chan events.Death
	GainExperience        <-chan events.GainExperience
	ItemAdded             <-chan events.ItemAdded
	PlayerFacilityCapture <-chan events.PlayerFacilityCapture
	PlayerFacilityDefend  <-chan events.PlayerFacilityDefend
	SkillAdded            <-chan events.SkillAdded
	VehicleDestroy        <-chan events.VehicleDestroy
	MetagameEvent         <-chan events.MetagameEvent
	FacilityControl       <-chan events.FacilityControl
	ContinentLock         <-chan events.ContinentLock
}

type eventSubscription struct {
	unsub func()
	close func()
}

func subscribeEvent[E events.Event](
	subs pubsub.SubscriptionsManager[events.EventType],
	ch chan E,
) eventSubscription {
	h := handler[events.EventType, E](ch)
	return eventSubscription{
		unsub: subs.AddHandler(h),
		close: func() { close(ch) },
	}
}

func SubscribeToEvents(
	postStopper module.Stopper,
	subs pubsub.SubscriptionsManager[events.EventType],
) EventsSubscriptionChannels {
	playerLogin := make(chan events.PlayerLogin)
	playerLogout := make(chan events.PlayerLogout)
	achievementEarned := make(chan events.AchievementEarned)
	battleRankUp := make(chan events.BattleRankUp)
	death := make(chan events.Death)
	gainExperience := make(chan events.GainExperience)
	itemAdded := make(chan events.ItemAdded)
	playerFacilityCapture := make(chan events.PlayerFacilityCapture)
	playerFacilityDefend := make(chan events.PlayerFacilityDefend)
	skillAdded := make(chan events.SkillAdded)
	vehicleDestroy := make(chan events.VehicleDestroy)
	metagameEvent := make(chan events.MetagameEvent)
	facilityControl := make(chan events.FacilityControl)
	continentLock := make(chan events.ContinentLock)

	subsList := []eventSubscription{
		subscribeEvent(subs, playerLogin),
		subscribeEvent(subs, playerLogout),
		subscribeEvent(subs, achievementEarned),
		subscribeEvent(subs, battleRankUp),
		subscribeEvent(subs, death),
		subscribeEvent(subs, gainExperience),
		subscribeEvent(subs, itemAdded),
		subscribeEvent(subs, playerFacilityCapture),
		subscribeEvent(subs, playerFacilityDefend),
		subscribeEvent(subs, skillAdded),
		subscribeEvent(subs, vehicleDestroy),
		subscribeEvent(subs, metagameEvent),
		subscribeEvent(subs, facilityControl),
		subscribeEvent(subs, continentLock),
	}

	postStopper.OnStop(module.NewRun(
		"events_subscription",
		func(_ context.Context) error {
			for _, s := range subsList {
				s.unsub()
				s.close()
			}
			return nil
		},
	))

	return EventsSubscriptionChannels{
		PlayerLogin:           playerLogin,
		PlayerLogout:          playerLogout,
		AchievementEarned:     achievementEarned,
		BattleRankUp:          battleRankUp,
		Death:                 death,
		GainExperience:        gainExperience,
		ItemAdded:             itemAdded,
		PlayerFacilityCapture: playerFacilityCapture,
		PlayerFacilityDefend:  playerFacilityDefend,
		SkillAdded:            skillAdded,
		VehicleDestroy:        vehicleDestroy,
		MetagameEvent:         metagameEvent,
		FacilityControl:       facilityControl,
		ContinentLock:         continentLock,
	}
}
