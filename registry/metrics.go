package registry

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	knownEndorsers   prometheus.Gauge
	temporalBanned   prometheus.Gauge
	permanentBanned  prometheus.Gauge
	trustedEndorsers prometheus.Gauge

	temporalBans     prometheus.Counter
	temporalForgives prometheus.Counter

	discoverRequests prometheus.Counter
	discoverSuccess  prometheus.Counter
	discoverCollided prometheus.Counter
	discoverFailures *prometheus.CounterVec

	askedForEndorser *prometheus.CounterVec

	discoverFailLowReputation   prometheus.Labels
	discoverFailNoReputation    prometheus.Labels
	discoverFailErrorReputation prometheus.Labels

	discoverTime prometheus.Histogram
}

func createMetrics(reg prometheus.Registerer) *metrics {
	knownEndorsers := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "registry_known_endorsers",
		Help: "Number of known endorsers",
	})

	temporalBanned := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "registry_temporal_banned",
		Help: "Number of temporarily banned endorsers",
	})

	permanentBanned := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "registry_permanent_banned",
		Help: "Number of permanently banned endorsers",
	})

	trustedEndorsers := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "registry_trusted_endorsers",
		Help: "Number of trusted endorsers",
	})

	temporalBans := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "registry_temporal_bans",
		Help: "Number of temporary bans",
	})

	temporalForgives := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "registry_temporal_forgives",
		Help: "Number of temporary ban forgives",
	})

	discoverRequests := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "registry_discover_requests",
		Help: "Number of discover requests",
	})

	discoverSuccess := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "registry_discover_success",
		Help: "Number of successful discovers",
	})

	discoverFailures := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "registry_discover_failures",
		Help: "Number of discover failures",
	}, []string{"reason"})

	discoverCollided := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "registry_discover_collided",
		Help: "Number of discover collisions",
	})

	discoverFailLowReputation := prometheus.Labels{
		"reason": "low_reputation",
	}

	discoverFailNoReputation := prometheus.Labels{
		"reason": "no_reputation",
	}

	discoverFailErrorReputation := prometheus.Labels{
		"reason": "error_reputation",
	}

	askedForEndorser := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "registry_asked_for_endorser",
		Help: "Number of times an endorser was asked for",
	}, []string{"status"})

	discoverTime := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "registry_discover_time",
		Help:    "Time taken to discover an endorser",
		Buckets: prometheus.DefBuckets,
	})

	if reg != nil {
		reg.MustRegister(
			knownEndorsers,
			temporalBanned,
			permanentBanned,
			trustedEndorsers,
			temporalBans,
			temporalForgives,
			discoverRequests,
			discoverSuccess,
			discoverFailures,
			discoverCollided,
			discoverTime,
			askedForEndorser,
		)
	}

	return &metrics{
		knownEndorsers:   knownEndorsers,
		temporalBanned:   temporalBanned,
		permanentBanned:  permanentBanned,
		trustedEndorsers: trustedEndorsers,

		temporalBans:     temporalBans,
		temporalForgives: temporalForgives,

		discoverRequests: discoverRequests,
		discoverSuccess:  discoverSuccess,
		discoverFailures: discoverFailures,
		discoverCollided: discoverCollided,

		askedForEndorser: askedForEndorser,

		discoverFailLowReputation:   discoverFailLowReputation,
		discoverFailNoReputation:    discoverFailNoReputation,
		discoverFailErrorReputation: discoverFailErrorReputation,

		discoverTime: discoverTime,
	}
}
