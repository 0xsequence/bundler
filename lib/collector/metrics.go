package collector

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	baseFee prometheus.Gauge

	failedFetchBaseFee   prometheus.Counter
	fetchBaseFeeDuration prometheus.Histogram

	minFeePerGas prometheus.GaugeVec
}

func createMetrics(reg prometheus.Registerer) *metrics {
	baseFee := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "collector_base_fee",
		Help: "Current base fee",
	})

	failedFetchBaseFee := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "collector_failed_fetch_base_fee",
		Help: "Number of failed base fee fetches",
	})

	fetchBaseFeeDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "collector_fetch_base_fee_duration",
		Help:    "Duration of fetching base fee",
		Buckets: prometheus.DefBuckets,
	})

	minFeePerGas := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "collector_min_fee_per_gas",
		Help: "Minimum fee per gas",
	}, []string{"token"})

	if reg != nil {
		reg.MustRegister(baseFee, failedFetchBaseFee, fetchBaseFeeDuration, minFeePerGas)
	}

	return &metrics{
		baseFee:              baseFee,
		failedFetchBaseFee:   failedFetchBaseFee,
		fetchBaseFeeDuration: fetchBaseFeeDuration,
		minFeePerGas:         *minFeePerGas,
	}
}
