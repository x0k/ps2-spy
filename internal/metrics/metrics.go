package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics interface {
}

type metrics struct {
	activePlayersQueueSize *prometheus.GaugeVec
	ps2EventsCount         *prometheus.CounterVec
}

func New(ns string) *metrics {
	return &metrics{
		activePlayersQueueSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: ns,
				Name:      "active_players_queue_size",
				Help:      "Active players queue size",
			},
			[]string{"platform"},
		),
		ps2EventsCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: ns,
				Name:      "ps2_events_count",
				Help:      "PS 2 events count",
			},
			[]string{"platform", "event_type"},
		),
	}
}

func (m *metrics) Register(reg prometheus.Registerer) {
	reg.MustRegister(m.activePlayersQueueSize)
}
