package metrics

import (
	"net/http"

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

type Status string

const (
	SuccessStatus Status = "ok"
	ErrorStatus   Status = "error"
)

type Subject string

const (
	RequestedSubject Subject = "requested"
	LoadedSubject    Subject = "loaded"
)

type PlatformLoaderName string

const (
	CharactersPlatformLoaderName PlatformLoaderName = "characters"
	CharacterPlatformLoaderName  PlatformLoaderName = "character"
)

type TransportName string

const (
	DefaultTransportName TransportName = "default"
)

type PlatformQueueName string

const (
	ActivePlayersQueueName PlatformQueueName = "active_players"
	LogoutEventsQueueName  PlatformQueueName = "logout_events"
)

type Metrics interface {
	PlatformLoadsCounterMetric(PlatformLoaderName, platforms.Platform) *prometheus.CounterVec
	PlatformLoaderInFlightMetric(PlatformLoaderName, platforms.Platform) *prometheus.Gauge
	PlatformLoaderSubjectsCounterMetric(PlatformLoaderName, platforms.Platform) *prometheus.CounterVec

	InstrumentPublisher(PublisherName, publisher.Publisher[publisher.Event]) publisher.Publisher[publisher.Event]
	InstrumentPlatformPublisher(PlatformPublisherName, platforms.Platform, publisher.Publisher[publisher.Event]) publisher.Publisher[publisher.Event]
	InstrumentTransport(TransportName, http.RoundTripper) http.RoundTripper

	SetPlatformQueueSize(PlatformQueueName, platforms.Platform, int)
}

type metrics struct {
	eventsCounter       *prometheus.CounterVec
	httpRequestsCounter *prometheus.CounterVec

	platformEventsCounter   *prometheus.CounterVec
	platformLoadsCounter    *prometheus.CounterVec
	platformLoadersInFlight *prometheus.GaugeVec
	platformLoadersSubjects *prometheus.CounterVec

	platformQueueSize *prometheus.GaugeVec
	platformCacheSize *prometheus.GaugeVec
}

func New(ns string) *metrics {
	return &metrics{
		eventsCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: ns,
				Name:      "events_count",
				Help:      "Events count",
			},
			[]string{"publisher_name", "event_type", "status"},
		),
		httpRequestsCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: ns,
				Name:      "http_requests_count",
				Help:      "HTTP requests count",
			},
			[]string{"transport_name", "host", "method", "status"},
		),
		platformEventsCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: ns,
				Name:      "platform_events_count",
				Help:      "Platform events count",
			},
			[]string{"publisher_name", "platform", "event_type", "status"},
		),
		platformLoadsCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: ns,
				Name:      "platform_loads_count",
				Help:      "Platform loads count",
			},
			[]string{"loader_name", "platform", "status"},
		),
		platformLoadersInFlight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: ns,
				Name:      "platform_loaders_in_flight",
				Help:      "Platform loaders in flight",
			},
			[]string{"loader_name", "platform"},
		),
		platformLoadersSubjects: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: ns,
				Name:      "platform_loaders_subjects",
				Help:      "Platform loaders subjects",
			},
			[]string{"loader_name", "platform", "subject"},
		),
		platformQueueSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: ns,
				Name:      "platform_queue_size",
				Help:      "Platform queue size",
			},
			[]string{"queue_name", "platform"},
		),
		platformCacheSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: ns,
				Name:      "platform_cache_size",
				Help:      "Platform cache size",
			},
			[]string{"cache_name", "platform"},
		),
	}
}

func (m *metrics) Register(reg prometheus.Registerer) {
	reg.MustRegister(
		m.eventsCounter,
		m.httpRequestsCounter,

		m.platformEventsCounter,
		m.platformLoadsCounter,
		m.platformLoadersInFlight,
		m.platformLoadersSubjects,

		m.platformQueueSize,
		m.platformCacheSize,
	)
}

func (m *metrics) InstrumentPublisher(
	name PublisherName,
	pub publisher.Publisher[publisher.Event],
) publisher.Publisher[publisher.Event] {
	return instrumentPublisher[publisher.Event](
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
	return instrumentPublisher[publisher.Event](
		m.platformEventsCounter.MustCurryWith(prometheus.Labels{
			"publisher_name": string(name),
			"platform":       string(platform),
		}),
		pub,
	)
}

func (m *metrics) InstrumentTransport(name TransportName, transport http.RoundTripper) http.RoundTripper {
	return instrumentTransport(
		m.httpRequestsCounter.MustCurryWith(prometheus.Labels{
			"transport_name": string(name),
		}),
		transport,
	)
}

func (m *metrics) PlatformLoadsCounterMetric(
	name PlatformLoaderName,
	platform platforms.Platform,
) *prometheus.CounterVec {
	return m.platformLoadsCounter.MustCurryWith(prometheus.Labels{
		"loader_name": string(name),
		"platform":    string(platform),
	})
}

func (m *metrics) PlatformLoaderInFlightMetric(
	name PlatformLoaderName,
	platform platforms.Platform,
) *prometheus.Gauge {
	g := m.platformLoadersInFlight.With(prometheus.Labels{
		"loader_name": string(name),
		"platform":    string(platform),
	})
	return &g
}

func (m *metrics) PlatformLoaderSubjectsCounterMetric(
	name PlatformLoaderName,
	platform platforms.Platform,
) *prometheus.CounterVec {
	return m.platformLoadersSubjects.MustCurryWith(prometheus.Labels{
		"loader_name": string(name),
		"platform":    string(platform),
	})
}

func (m *metrics) SetPlatformQueueSize(
	name PlatformQueueName,
	platform platforms.Platform,
	size int,
) {
	m.platformQueueSize.With(prometheus.Labels{
		"queue_name": string(name),
		"platform":   string(platform),
	}).Set(float64(size))
}
