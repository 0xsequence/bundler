package utils

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func RecordFunctionDuration(start time.Time, histogram prometheus.Histogram) {
	// Calculate the duration since 'start', convert to seconds.
	duration := time.Since(start).Seconds()

	// Record the duration into the histogram.
	histogram.Observe(duration)
}

func RecordFunctionDurationVec(start time.Time, histogram *prometheus.HistogramVec, labels prometheus.Labels) {
	// Calculate the duration since 'start', convert to seconds.
	duration := time.Since(start).Seconds()

	// Record the duration into the histogram.
	histogram.With(labels).Observe(duration)
}
