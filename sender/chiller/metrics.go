package chiller

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	chilledOps prometheus.Gauge
	blockedOps prometheus.Counter
}

func createMetrics() *metrics {
	return &metrics{
		chilledOps: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "sender_chilled_ops",
			Help: "Number of operations that are currently chilled",
		}),
		blockedOps: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sender_blocked_ops",
			Help: "Number of operations that are currently blocked",
		}),
	}
}

func (m *metrics) register(reg prometheus.Registerer) {
	reg.MustRegister(
		m.chilledOps,
		m.blockedOps,
	)
}
