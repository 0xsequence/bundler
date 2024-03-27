package sender

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	selectOpsTime prometheus.Histogram

	skipRunNoOps prometheus.Counter
}

func createMetrics() *metrics {
	return &metrics{
		selectOpsTime: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "sender_select_ops_time",
			Help:    "Time it takes to select operations",
			Buckets: prometheus.ExponentialBuckets(1e6, 2, 32),
		}),

		skipRunNoOps: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sender_skip_run_no_ops",
			Help: "Number of times the sender skipped running because there were no operations",
		}),
	}
}

func (m *metrics) register(reg prometheus.Registerer) {
	reg.MustRegister(
		m.selectOpsTime,
		m.skipRunNoOps,
	)
}
