package worker

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	attemptSendOps   prometheus.Counter
	executedOps      prometheus.Counter
	failedReceiptOps prometheus.Counter
	failedSendOps    *prometheus.CounterVec

	preparedAndDroppedOps prometheus.Counter

	failedSimulateOperation prometheus.Labels
	failedEndorserLied      prometheus.Labels
	failedStaleOperation    prometheus.Labels
	failedEstimateGas       prometheus.Labels
	failedSignTransaction   prometheus.Labels
	failedSendTransaction   prometheus.Labels

	inspectReceiptReverted *prometheus.CounterVec
	inspectReceiptFailed   *prometheus.CounterVec

	failedInspectReceiptBalanceOf1        prometheus.Labels
	failedInspectReceiptBalanceOf2        prometheus.Labels
	failedInspectReceiptEffectiveGasPrice prometheus.Labels

	sendOpTime         prometheus.Histogram
	prepareOpTime      prometheus.Histogram
	waitReceiptTime    prometheus.Histogram
	inspectReceiptTime prometheus.Histogram
	simulateOpTime     prometheus.Histogram

	overpaidAmount  prometheus.Histogram
	underpaidAmount prometheus.Histogram

	profitableOpDiff   prometheus.Histogram
	unprofitableOpDiff prometheus.Histogram
}

func createMetrics() *metrics {
	return &metrics{
		attemptSendOps: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sender_attempt_send_ops",
			Help: "Number of attempted send operations",
		}),
		executedOps: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sender_executed_ops",
			Help: "Number of executed operations",
		}),
		failedReceiptOps: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sender_failed_receipt_ops",
			Help: "Number of times the sender failed to wait for a receipt",
		}),
		failedSendOps: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "sender_failed_send_ops",
			Help: "Number of failed send operations",
		}, []string{"reason"}),
		failedSimulateOperation: prometheus.Labels{"reason": "simulate_operation"},
		failedEndorserLied:      prometheus.Labels{"reason": "endorser_lied"},
		failedStaleOperation:    prometheus.Labels{"reason": "stale_operation"},
		failedEstimateGas:       prometheus.Labels{"reason": "estimate_gas"},
		failedSignTransaction:   prometheus.Labels{"reason": "sign_transaction"},
		failedSendTransaction:   prometheus.Labels{"reason": "send_transaction"},
		inspectReceiptReverted: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "sender_inspect_receipt_reverted",
			Help: "Number of inspect receipt reverted",
		}, []string{"lied"}),
		inspectReceiptFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "sender_inspect_receipt_failed",
			Help: "Number of inspect receipt failed",
		}, []string{"reason"}),
		failedInspectReceiptBalanceOf1:        prometheus.Labels{"reason": "balance_of_1"},
		failedInspectReceiptBalanceOf2:        prometheus.Labels{"reason": "balance_of_2"},
		failedInspectReceiptEffectiveGasPrice: prometheus.Labels{"reason": "effective_gas_price"},
		sendOpTime: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "sender_send_op_time",
			Help:    "Time it takes to send an operation",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 13),
		}),
		prepareOpTime: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "sender_prepare_op_time",
			Help:    "Time it takes to prepare an operation",
			Buckets: prometheus.DefBuckets,
		}),
		waitReceiptTime: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "sender_wait_receipt_time",
			Help:    "Time it takes to wait for a receipt",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 13),
		}),
		inspectReceiptTime: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "sender_inspect_receipt_time",
			Help:    "Time it takes to inspect a receipt",
			Buckets: prometheus.DefBuckets,
		}),
		simulateOpTime: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "sender_simulate_op_time",
			Help:    "Time it takes to simulate an operation",
			Buckets: prometheus.DefBuckets,
		}),
		overpaidAmount: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "sender_overpaid_amount",
			Help:    "Amount overpaid in native token",
			Buckets: prometheus.ExponentialBuckets(1000000000, 2, 32),
		}),
		underpaidAmount: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "sender_underpaid_amount",
			Help:    "Amount underpaid in native token",
			Buckets: prometheus.ExponentialBuckets(1000000000, 2, 32),
		}),
		profitableOpDiff: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "sender_profitable_op_diff",
			Help:    "Difference between expected and required payment in native token",
			Buckets: prometheus.ExponentialBuckets(1000000000, 2, 32),
		}),
		unprofitableOpDiff: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "sender_unprofitable_op_diff",
			Help:    "Difference between expected and required payment in native token",
			Buckets: prometheus.ExponentialBuckets(1000000000, 2, 32),
		}),
		preparedAndDroppedOps: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sender_prepared_and_dropped_ops_count",
			Help: "Number of prepared operations that were dropped after being prepared",
		}),
	}
}

func (m *metrics) register(reg prometheus.Registerer) {
	reg.MustRegister(
		m.attemptSendOps,
		m.executedOps,
		m.failedReceiptOps,
		m.failedSendOps,
		m.inspectReceiptReverted,
		m.inspectReceiptFailed,
		m.sendOpTime,
		m.prepareOpTime,
		m.waitReceiptTime,
		m.inspectReceiptTime,
		m.simulateOpTime,
		m.overpaidAmount,
		m.underpaidAmount,
		m.profitableOpDiff,
		m.unprofitableOpDiff,
		m.preparedAndDroppedOps,
	)
}
