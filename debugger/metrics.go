package debugger

import "github.com/prometheus/client_golang/prometheus"

type anvilMetrics struct {
	startAttempts  prometheus.Counter
	startSuccesses prometheus.Counter
	startFailures  prometheus.Counter

	stopOperations  prometheus.Counter
	resetOperations prometheus.Counter

	debugTraceCallOperations prometheus.Counter
	debugTraceCallRetry      prometheus.Counter
	debugTraceCallSuccesses  prometheus.Counter
	debugTraceCallFailures   prometheus.Counter

	anvilRunning prometheus.Gauge

	ipcWaitFailures       *prometheus.CounterVec
	ipcWaitFailureError   prometheus.Labels
	ipcWaitFailureTimeout prometheus.Labels

	debugCallDuration   prometheus.Histogram
	ipcFileWaitDuration prometheus.Histogram
}

func createAnvilMetrics(reg prometheus.Registerer) *anvilMetrics {
	startAttempts := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "anvil_start_attempts",
		Help: "Number of attempts to start anvil",
	})

	startSuccesses := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "anvil_start_successes",
		Help: "Number of successful starts of anvil",
	})

	startFailures := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "anvil_start_failures",
		Help: "Number of failed starts of anvil",
	})

	stopOperations := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "anvil_stop_operations",
		Help: "Number of stop operations on anvil",
	})

	resetOperations := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "anvil_reset_operations",
		Help: "Number of reset operations on anvil",
	})

	debugTraceCallOperations := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "anvil_debug_trace_call_operations",
		Help: "Number of debug trace call operations on anvil",
	})

	debugTraceCallSuccesses := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "anvil_debug_trace_call_successes",
		Help: "Number of successful debug trace call operations on anvil",
	})

	debugTraceCallRetry := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "anvil_debug_trace_call_retry",
		Help: "Number of retries for debug trace call operations on anvil",
	})

	debugTraceCallFailures := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "anvil_debug_trace_call_failures",
		Help: "Number of failed debug trace call operations on anvil",
	})

	anvilRunning := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "anvil_running",
		Help: "Anvil running state",
	})

	debugCallDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "anvil_debug_call_duration",
		Help: "Duration of anvil debug call operations",
		Buckets: []float64{
			0.25, // 0.25 seconds
			0.5,  // 0.5 seconds
			1,    // 1 second
			2.5,  // 2.5 seconds
			5,    // 5 seconds
			10,   // 10 seconds
			15,   // 15 seconds
			30,   // 30 seconds
			45,   // 45 seconds
			60,   // 60 seconds
		},
	})

	ipcWaitFailures := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "anvil_ipc_wait_failures",
		Help: "Number of failures waiting for anvil ipc file",
	}, []string{"reason"})

	ipcFileWaitDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "anvil_ipc_file_wait_duration",
		Help:    "Duration of waiting for anvil ipc file",
		Buckets: prometheus.DefBuckets,
	})

	if reg != nil {
		reg.MustRegister(
			startAttempts, startSuccesses, startFailures,
			stopOperations, resetOperations,
			debugTraceCallOperations, debugTraceCallSuccesses,
			debugTraceCallRetry,
			ipcWaitFailures, debugTraceCallFailures,
			anvilRunning, debugCallDuration, ipcFileWaitDuration,
		)
	}

	return &anvilMetrics{
		startAttempts:            startAttempts,
		startSuccesses:           startSuccesses,
		startFailures:            startFailures,
		stopOperations:           stopOperations,
		resetOperations:          resetOperations,
		debugTraceCallOperations: debugTraceCallOperations,
		debugTraceCallSuccesses:  debugTraceCallSuccesses,
		debugTraceCallRetry:      debugTraceCallRetry,
		debugTraceCallFailures:   debugTraceCallFailures,
		anvilRunning:             anvilRunning,
		debugCallDuration:        debugCallDuration,
		ipcWaitFailures:          ipcWaitFailures,
		ipcFileWaitDuration:      ipcFileWaitDuration,

		ipcWaitFailureError:   prometheus.Labels{"reason": "error"},
		ipcWaitFailureTimeout: prometheus.Labels{"reason": "timeout"},
	}
}
