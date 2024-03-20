package partitioner

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	overlapLimit  prometheus.Gauge
	wildcardLimit prometheus.Gauge

	knownOps prometheus.Gauge

	wildcardDeps        prometheus.Gauge
	partiallyFilledDeps prometheus.Gauge
	fullDependencies    prometheus.Gauge

	addTimes          prometheus.Histogram
	removeTimes       prometheus.Histogram
	overlapCollisions prometheus.Histogram

	wildcardCollisions prometheus.Counter
	knownCollisions    prometheus.Counter

	addedDependencies   prometheus.Counter
	removedDependencies prometheus.Counter
}

func createMetrics(reg prometheus.Registerer, overlapLimit, wildcardLimit uint) *metrics {
	m := &metrics{
		overlapLimit:  prometheus.NewGauge(prometheus.GaugeOpts{Name: "mempool_partitioner_overlap_limit"}),
		wildcardLimit: prometheus.NewGauge(prometheus.GaugeOpts{Name: "mempool_partitioner_wildcard_limit"}),

		knownOps: prometheus.NewGauge(prometheus.GaugeOpts{Name: "mempool_partitioner_known_ops"}),

		wildcardDeps:        prometheus.NewGauge(prometheus.GaugeOpts{Name: "mempool_partitioner_wildcard_dependencies"}),
		partiallyFilledDeps: prometheus.NewGauge(prometheus.GaugeOpts{Name: "mempool_partitioner_partially_filled_dependencies"}),
		fullDependencies:    prometheus.NewGauge(prometheus.GaugeOpts{Name: "mempool_partitioner_full_dependencies"}),

		addTimes: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "mempool_partitioner_add_times",
			Buckets: prometheus.ExponentialBuckets(1e-6, 2, 15),
		}),

		removeTimes: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "mempool_partitioner_remove_times",
			Buckets: prometheus.ExponentialBuckets(1e-6, 2, 15),
		}),

		overlapCollisions: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "mempool_partitioner_overlap_collisions",
			Buckets: prometheus.LinearBuckets(0, float64(overlapLimit), 20),
		}),

		wildcardCollisions: prometheus.NewCounter(prometheus.CounterOpts{Name: "mempool_partitioner_wildcard_collisions"}),
		knownCollisions:    prometheus.NewCounter(prometheus.CounterOpts{Name: "mempool_partitioner_known_collisions"}),

		addedDependencies:   prometheus.NewCounter(prometheus.CounterOpts{Name: "mempool_partitioner_added_dependencies"}),
		removedDependencies: prometheus.NewCounter(prometheus.CounterOpts{Name: "mempool_partitioner_removed_dependencies"}),
	}

	if reg != nil {
		reg.MustRegister(
			m.overlapLimit, m.wildcardLimit,
			m.knownOps,
			m.wildcardDeps, m.partiallyFilledDeps, m.fullDependencies,
			m.addTimes, m.removeTimes, m.overlapCollisions,
			m.addedDependencies, m.removedDependencies,
			m.wildcardCollisions, m.knownCollisions,
		)
	}

	m.overlapLimit.Set(float64(overlapLimit))
	m.wildcardLimit.Set(float64(wildcardLimit))

	return m
}
