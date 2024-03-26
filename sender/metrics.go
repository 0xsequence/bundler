package sender

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	attemptSendOps   prometheus.Counter
	unprofitableOps  prometheus.Counter
	executedOps      prometheus.Counter
	failedReceiptOps prometheus.Counter
	failedSendOps    *prometheus.CounterVec

	failedSimulateOperation prometheus.Labels
	failedEndorserLied      prometheus.Labels
	failedStaleOperation    prometheus.Labels
	failedEstimateGas       prometheus.Labels
	failedSignTransaction   prometheus.Labels
	failedSendTransaction   prometheus.Labels

	chilledOps prometheus.Gauge
	blockedOps prometheus.Counter

	inspectReceiptAttempts  prometheus.Counter
	inspectReceiptPaid      prometheus.Counter
	inspectReceiptUnderpaid prometheus.Counter
	inspectReceiptReverted  *prometheus.CounterVec
	inspectReceiptFailed    *prometheus.CounterVec

	failedInspectReceiptBalanceOf1        prometheus.Labels
	failedInspectReceiptBalanceOf2        prometheus.Labels
	failedInspectReceiptEffectiveGasPrice prometheus.Labels

	sendOpTime         prometheus.Histogram
	prepareOpTime      prometheus.Histogram
	waitReceiptTime    prometheus.Histogram
	inspectReceiptTime prometheus.Histogram
	simulateOpTime     prometheus.Histogram
	selectOpsTime      prometheus.Histogram

	overpaidAmount  prometheus.Histogram
	underpaidAmount prometheus.Histogram

	profitableOpDiff   prometheus.Histogram
	unprofitableOpDiff prometheus.Histogram

	skipRunNoOps prometheus.Counter
}

func createMetrics(reg prometheus.Registerer, sender string) *metrics {
	attemptSendOps := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "sender_attempt_send_ops",
		Help: "Number of attempted send operations",
	})

	unprofitableOps := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "sender_unprofitable_ops",
		Help: "Number of unprofitable operations",
	})

	executedOps := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "sender_executed_ops",
		Help: "Number of executed operations",
	})

	failedReceiptOps := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "sender_failed_receipt_ops",
		Help: "Number of times the sender failed to wait for a receipt",
	})

	failedSendOps := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "sender_failed_send_ops",
		Help: "Number of failed send operations",
	}, []string{"reason"})

	chilledOps := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "sender_chilled_ops",
		Help: "Number of chilled operations (waiting for better payment)",
	})

	blockedOps := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "sender_blocked_ops",
		Help: "Number of blocked operations (already executed)",
	})

	inspectReceiptAttempts := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "sender_inspect_receipt_attempts",
		Help: "Number of inspect receipt attempts",
	})

	inspectReceiptPaid := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "sender_inspect_receipt_paid",
		Help: "Number of inspect receipt paid",
	})

	inspectReceiptUnderpaid := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "sender_inspect_receipt_underpaid",
		Help: "Number of inspect receipt underpaid",
	})

	inspectReceiptReverted := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "sender_inspect_receipt_reverted",
		Help: "Number of inspect receipt reverted",
	}, []string{"lied"})

	inspectReceiptFailed := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "sender_inspect_receipt_failed",
		Help: "Number of inspect receipt failed",
	}, []string{"reason"})

	sendOpTime := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "sender_send_op_time",
		Help:    "Time it takes to send an operation",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 13),
	})

	prepareOpTime := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "sender_prepare_op_time",
		Help:    "Time it takes to prepare an operation",
		Buckets: prometheus.DefBuckets,
	})

	waitReceiptTime := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "sender_wait_receipt_time",
		Help:    "Time it takes to wait for a receipt",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 13),
	})

	inspectReceiptTime := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "sender_inspect_receipt_time",
		Help:    "Time it takes to inspect a receipt",
		Buckets: prometheus.DefBuckets,
	})

	simulateOpTime := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "sender_simulate_op_time",
		Help:    "Time it takes to simulate an operation",
		Buckets: prometheus.DefBuckets,
	})

	selectOpsTime := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "sender_select_ops_time",
		Help:    "Time it takes to select operations",
		Buckets: prometheus.ExponentialBuckets(1e6, 2, 32),
	})

	overpaidAmount := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "sender_overpaid_amount",
		Help:    "Amount overpaid in native token",
		Buckets: prometheus.ExponentialBuckets(1000000000, 2, 32),
	})

	underpaidAmount := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "sender_underpaid_amount",
		Help:    "Amount underpaid in native token",
		Buckets: prometheus.ExponentialBuckets(1000000000, 2, 32),
	})

	unprofitableOpDiff := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "sender_unprofitable_op_diff",
		Help:    "Difference between expected and required payment in native token",
		Buckets: prometheus.ExponentialBuckets(1000000000, 2, 32),
	})

	profitableOpDiff := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "sender_profitable_op_diff",
		Help:    "Difference between expected and required payment in native token",
		Buckets: prometheus.ExponentialBuckets(1000000000, 2, 32),
	})

	skipRunNoOps := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "sender_skip_run_no_ops",
		Help: "Number of times the sender skipped running because there were no operations",
	})

	if reg != nil {
		regWithSender := prometheus.WrapRegistererWith(prometheus.Labels{"sender": sender}, reg)
		regWithSender.MustRegister(
			attemptSendOps,
			unprofitableOps,
			executedOps,
			failedReceiptOps,
			failedSendOps,
			chilledOps,
			blockedOps,
			inspectReceiptAttempts,
			inspectReceiptPaid,
			inspectReceiptUnderpaid,
			inspectReceiptReverted,
			inspectReceiptFailed,
			sendOpTime,
			prepareOpTime,
			waitReceiptTime,
			inspectReceiptTime,
			simulateOpTime,
			overpaidAmount,
			underpaidAmount,
			unprofitableOpDiff,
			profitableOpDiff,
			skipRunNoOps,
			selectOpsTime,
		)
	}

	return &metrics{
		attemptSendOps:   attemptSendOps,
		unprofitableOps:  unprofitableOps,
		executedOps:      executedOps,
		failedReceiptOps: failedReceiptOps,
		failedSendOps:    failedSendOps,

		failedSimulateOperation: prometheus.Labels{"reason": "simulate_operation"},
		failedEndorserLied:      prometheus.Labels{"reason": "endorser_lied"},
		failedStaleOperation:    prometheus.Labels{"reason": "stale_operation"},
		failedEstimateGas:       prometheus.Labels{"reason": "estimate_gas"},
		failedSignTransaction:   prometheus.Labels{"reason": "sign_transaction"},
		failedSendTransaction:   prometheus.Labels{"reason": "send_transaction"},

		chilledOps: chilledOps,
		blockedOps: blockedOps,

		inspectReceiptAttempts:  inspectReceiptAttempts,
		inspectReceiptPaid:      inspectReceiptPaid,
		inspectReceiptUnderpaid: inspectReceiptUnderpaid,
		inspectReceiptReverted:  inspectReceiptReverted,
		inspectReceiptFailed:    inspectReceiptFailed,

		failedInspectReceiptBalanceOf1:        prometheus.Labels{"reason": "balance_of_1"},
		failedInspectReceiptBalanceOf2:        prometheus.Labels{"reason": "balance_of_2"},
		failedInspectReceiptEffectiveGasPrice: prometheus.Labels{"reason": "effective_gas_price"},

		sendOpTime:         sendOpTime,
		prepareOpTime:      prepareOpTime,
		waitReceiptTime:    waitReceiptTime,
		inspectReceiptTime: inspectReceiptTime,
		simulateOpTime:     simulateOpTime,
		selectOpsTime:      selectOpsTime,

		overpaidAmount:     overpaidAmount,
		underpaidAmount:    underpaidAmount,
		profitableOpDiff:   profitableOpDiff,
		unprofitableOpDiff: unprofitableOpDiff,

		skipRunNoOps: skipRunNoOps,
	}
}
