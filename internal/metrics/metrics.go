package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type PublisherName string

const (
	StoragePublisher PublisherName = "storage"
)

type PlatformPublisherName string

const (
	Ps2MessagesPlatformPublisher         PlatformPublisherName = "ps2_messages"
	Ps2EventsPlatformPublisher           PlatformPublisherName = "ps2_events"
	CharactersTrackerPlatformPublisher   PlatformPublisherName = "characters_tracker"
	WorldsTrackerPlatformPublisher       PlatformPublisherName = "worlds_tracker"
	OutfitsMembersSaverPlatformPublisher PlatformPublisherName = "outfits_members_saver"
)

type Metrics interface {
	InstrumentPublisher(PublisherName, publisher.Publisher[publisher.Event]) publisher.Publisher[publisher.Event]
	InstrumentPlatformPublisher(PlatformPublisherName, platforms.Platform, publisher.Publisher[publisher.Event]) publisher.Publisher[publisher.Event]
}

type metrics struct {
	eventsCounter         *prometheus.CounterVec
	platformEventsCounter *prometheus.CounterVec
	platformQueueSize     *prometheus.GaugeVec
	platformBatchSize     *prometheus.GaugeVec
	platformCacheSize     *prometheus.GaugeVec
}

func New(ns string) *metrics {
	return &metrics{
		eventsCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: ns,
				Name:      "events_count",
				Help:      "Events count",
			},
			[]string{"publisher_name", "event_type"},
		),
		platformEventsCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: ns,
				Name:      "platform_events_count",
				Help:      "Platform events count",
			},
			[]string{"publisher_name", "platform", "event_type"},
		),
		platformQueueSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: ns,
				Name:      "queue_size",
				Help:      "App queue size",
			},
			[]string{"queue_name", "platform"},
		),
		platformBatchSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: ns,
				Name:      "batch_size",
				Help:      "Batch size",
			},
			[]string{"batcher_name", "platform"},
		),
		platformCacheSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: ns,
				Name:      "cache_size",
				Help:      "Cache size",
			},
			[]string{"cache_name", "platform"},
		),
	}
}

func (m *metrics) Register(reg prometheus.Registerer) {
	reg.MustRegister(
		m.platformEventsCounter,
		m.platformQueueSize,
		m.platformBatchSize,
		m.platformCacheSize,
	)
}

func (m *metrics) InstrumentPublisher(
	name PublisherName,
	pub publisher.Publisher[publisher.Event],
) publisher.Publisher[publisher.Event] {
	return instrumentPublisherCounter[publisher.Event](
		m.eventsCounter.MustCurryWith(prometheus.Labels{
			"publisher_name": string(name),
		}),
		pub,
	)
}

func (m *metrics) InstrumentPlatformPublisher(
	name PlatformPublisherName,
	platform platforms.Platform,
	pub publisher.Publisher[publisher.Event],
) publisher.Publisher[publisher.Event] {
	return instrumentPublisherCounter[publisher.Event](
		m.platformEventsCounter.MustCurryWith(prometheus.Labels{
			"publisher_name": string(name),
			"platform":       string(platform),
		}),
		pub,
	)
}
