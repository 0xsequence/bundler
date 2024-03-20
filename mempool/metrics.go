package mempool

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	ops   prometheus.Gauge
	known prometheus.Gauge

	opsRejected        *prometheus.CounterVec
	opsBroadcastFailed prometheus.Counter

	opRejectedKnown              prometheus.Labels
	opRejectedReadyErr           prometheus.Labels
	opRejectedReadyNotReady      prometheus.Labels
	opRejectedConstraintsErr     prometheus.Labels
	opRejectedConstraintsNotMet  prometheus.Labels
	opRejectedDependencyStateErr prometheus.Labels
	opRejectedCollectorErr       prometheus.Labels
	opRejectedRegistryErr        prometheus.Labels
	opRejectedNoEviction         prometheus.Labels
	opRejectedNoEvictionGlobal   prometheus.Labels
	opRejectedPartitionerRace    prometheus.Labels

	opsEvicted   prometheus.Counter
	opsDiscarded prometheus.Counter

	opsMarkedForForget prometheus.Counter
	opsForgotten       prometheus.Counter

	opsReserved prometheus.Counter
	opsReleased *prometheus.CounterVec

	opAddedTime      prometheus.Histogram
	opLifetime       prometheus.Histogram
	reservedTime     prometheus.Histogram
	doReserveOpsTime prometheus.Histogram
}

func createMetrics(reg prometheus.Registerer) *metrics {
	ops := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "mempool_ops",
		Help: "Number of operations in the mempool",
	})

	known := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "mempool_known",
		Help: "Number of known operations",
	})

	reserved := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "mempool_reserved",
		Help: "Number of reserved operations",
	})

	opsRejected := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "mempool_ops_rejected",
		Help: "Number of operations rejected",
	}, []string{"reason"})

	opsBroadcastFailed := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mempool_ops_broadcast_failed",
		Help: "Number of failed operation broadcasts",
	})

	opsEvicted := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mempool_ops_evicted",
		Help: "Number of operations evicted",
	})

	opsDiscarded := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mempool_ops_discarded",
		Help: "Number of operations discarded",
	})

	opsMarkedForForget := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mempool_ops_marked_for_forget",
		Help: "Number of operations marked for forget",
	})

	opsForgotten := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mempool_ops_forgotten",
		Help: "Number of operations forgotten",
	})

	opsReserved := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mempool_ops_reserved",
		Help: "Number of operations reserved",
	})

	opsReleased := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "mempool_ops_released",
		Help: "Number of operations released",
	}, []string{"change"})

	opAddedTime := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "mempool_op_added_time",
		Help:    "Time it takes to add an operation to the mempool",
		Buckets: prometheus.DefBuckets,
	})

	opLifetime := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "mempool_op_lifetime",
		Help:    "Lifetime of an operation in the mempool",
		Buckets: prometheus.ExponentialBuckets(1, 2, 25),
	})

	opsReservedTime := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "mempool_ops_reserved_time",
		Help:    "Time an operation is reserved",
		Buckets: prometheus.ExponentialBuckets(1, 2, 25),
	})

	doReserveOpsTime := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "mempool_do_reserve_ops_time",
		Help:    "Time it takes to reserve operations",
		Buckets: prometheus.ExponentialBuckets(1e-6, 2, 15),
	})

	if reg != nil {
		reg.MustRegister(
			ops,
			known,
			reserved,
			opsRejected,
			opsBroadcastFailed,
			opsEvicted,
			opsDiscarded,
			opsMarkedForForget,
			opsForgotten,
			opsReserved,
			opsReleased,
			opAddedTime,
			opLifetime,
			opsReservedTime,
			doReserveOpsTime,
		)
	}

	return &metrics{
		ops:   ops,
		known: known,

		opsRejected:        opsRejected,
		opsBroadcastFailed: opsBroadcastFailed,

		opRejectedKnown:              prometheus.Labels{"reason": "known"},
		opRejectedReadyErr:           prometheus.Labels{"reason": "ready_err"},
		opRejectedReadyNotReady:      prometheus.Labels{"reason": "ready_not_ready"},
		opRejectedConstraintsErr:     prometheus.Labels{"reason": "constraints_err"},
		opRejectedConstraintsNotMet:  prometheus.Labels{"reason": "constraints_not_met"},
		opRejectedDependencyStateErr: prometheus.Labels{"reason": "dependency_state_err"},
		opRejectedCollectorErr:       prometheus.Labels{"reason": "collector_err"},
		opRejectedRegistryErr:        prometheus.Labels{"reason": "registry_err"},
		opRejectedNoEviction:         prometheus.Labels{"reason": "no_eviction"},
		opRejectedNoEvictionGlobal:   prometheus.Labels{"reason": "no_eviction_global"},
		opRejectedPartitionerRace:    prometheus.Labels{"reason": "partition_race"},

		opsEvicted:   opsEvicted,
		opsDiscarded: opsDiscarded,

		opsMarkedForForget: opsMarkedForForget,
		opsForgotten:       opsForgotten,

		opsReserved: opsReserved,
		opsReleased: opsReleased,

		opAddedTime:      opAddedTime,
		opLifetime:       opLifetime,
		reservedTime:     opsReservedTime,
		doReserveOpsTime: doReserveOpsTime,
	}
}
