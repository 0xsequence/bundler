package ipfs

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	ReportedTime  prometheus.Histogram
	ReportedBytes prometheus.Histogram

	ReportFailed prometheus.Counter
}

func createMetrics(reg prometheus.Registerer) *metrics {
	reportedTime := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "ipfs_reported_time",
		Help:    "Time taken to report data to IPFS",
		Buckets: prometheus.DefBuckets,
	})

	ReportedBytes := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "ipfs_reported_bytes",
		Help:    "Weight of data reported to IPFS",
		Buckets: prometheus.ExponentialBucketsRange(32, 65536, 12),
	})

	reportFailed := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ipfs_report_failed",
		Help: "Number of failed reports to IPFS",
	})

	if reg != nil {
		reg.MustRegister(reportedTime, ReportedBytes, reportFailed)
	}

	return &metrics{
		ReportedTime:  reportedTime,
		ReportedBytes: ReportedBytes,
		ReportFailed:  reportFailed,
	}
}
