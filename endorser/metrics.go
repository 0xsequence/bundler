package endorser

import (
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	isOperationReadyAttempts       prometheus.Counter
	isOperationReadyWildcards      prometheus.Counter
	isOperationReadyDebugger       prometheus.Counter
	isOperationReadyDebuggerFailed prometheus.Counter
	isOperationReadyError          prometheus.Counter
	isOperationReadyTrue           prometheus.Counter
	isOperationReadyFalse          prometheus.Counter
	isOperationReadyReverts        prometheus.Counter

	failoverSimulationAttempts prometheus.Counter
	failoverSimulationSuccess  prometheus.Counter
	failoverSimulationError    prometheus.Counter

	isOperationReadyDuration      prometheus.Histogram
	isOperationDebugReadyDuration prometheus.Histogram
	failoverSimulationDuration    prometheus.Histogram

	durationPerGas      prometheus.Histogram
	debugDurationPerGas prometheus.Histogram

	dependencyStateDuration prometheus.Histogram
	constraintsMetDuration  prometheus.Histogram
	constraintMetDuration   prometheus.Histogram
	dependencySlotDuration  prometheus.Histogram

	constraintsMet       prometheus.Counter
	constraintsNotMet    prometheus.Counter
	dependencyStateError *prometheus.CounterVec
	constraintsMetError  prometheus.Counter

	dependencyStateErrorBalance prometheus.Labels
	dependencyStateErrorCode    prometheus.Labels
	dependencyStateErrorNonce   prometheus.Labels
	dependencyStateErrorSlots   prometheus.Labels
}

func createMetrics(reg prometheus.Registerer) *metrics {
	isOperationReadyAttempts := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "endorser_is_operation_ready_attempts",
		Help: "Number of attempts to check if an operation is ready",
	})

	isOperationReadyWildcards := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "endorser_is_operation_ready_wildcards_sum",
		Help: "Number of attempts to check if an operation that resulted in wildcards",
	})

	isOperationReadyDebugger := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "endorser_is_operation_ready_debugger",
		Help: "Number of attempts to check if an operation is ready using the debugger",
	})

	isOperationReadyDebuggerFailed := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "endorser_is_operation_ready_debugger_failed",
		Help: "Number of failed attempts to check if an operation is ready using the debugger",
	})

	isOperationReadyError := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "endorser_is_operation_ready_error",
		Help: "Number of errors when checking if an operation is ready",
	})

	isOperationReadyTrue := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "endorser_is_operation_ready_true_sum",
		Help: "Number of operations with readiness true",
	})

	isOperationReadyFalse := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "endorser_is_operation_ready_false_sum",
		Help: "Number of operations with readiness false",
	})

	isOperationReadyReverts := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "endorser_is_operation_ready_reverts",
		Help: "Number of operations that reverted when checking if they are ready",
	})

	failoverSimulationAttempts := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "endorser_failover_simulation_attempts",
		Help: "Number of failover attempts to simulate an operation",
	})

	failoverSimulationSuccess := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "endorser_failover_simulation_success",
		Help: "Number of successful failover attempts to simulate an operation",
	})

	failoverSimulationError := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "endorser_failover_simulation_error",
		Help: "Number of failed failover attempts to simulate an operation",
	})

	isOperationReadyDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "endorser_is_operation_ready_duration",
		Help:    "Duration to check if an operation is ready",
		Buckets: prometheus.DefBuckets,
	})

	isOperationDebugReadyDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "endorser_is_operation_debug_ready_duration",
		Help:    "Duration to check if an operation is ready using the debugger",
		Buckets: prometheus.DefBuckets,
	})

	failoverSimulationDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "endorser_failover_simulation_duration",
		Help:    "Duration to check an operation using failover simulation",
		Buckets: prometheus.DefBuckets,
	})

	durationPerGas := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "endorser_duration_per_gas",
		Help:    "Duration per gas",
		Buckets: prometheus.DefBuckets,
	})

	debugDurationPerGas := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "endorser_debug_duration_per_gas",
		Help:    "Duration per gas using the debugger",
		Buckets: prometheus.DefBuckets,
	})

	dependencyStateDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "endorser_dependency_state_duration",
		Help:    "Duration to get the state of dependencies",
		Buckets: prometheus.DefBuckets,
	})

	constraintsMetDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "endorser_constraints_met_duration",
		Help:    "Duration to check if constraints are met",
		Buckets: prometheus.DefBuckets,
	})

	constraintsMet := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "endorser_constraints_met_sum",
		Help: "Number of constraints met",
	})

	constraintsNotMet := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "endorser_constraints_not_met",
		Help: "Number of constraints not met",
	})

	dependencyStateError := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "endorser_dependency_state_error_sum",
		Help: "Number of errors when getting the state of dependencies",
	}, []string{"reason"})

	constraintsMetError := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "endorser_constraints_met_error_sum",
		Help: "Number of errors when checking if constraints are met",
	})

	constraintMetDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "endorser_constraint_met_duration",
		Help:    "Duration to check if one constraint is met",
		Buckets: prometheus.DefBuckets,
	})

	dependencySlotDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "endorser_dependency_slot_duration",
		Help:    "Duration to get the state of a single slot",
		Buckets: prometheus.DefBuckets,
	})

	if reg != nil {
		reg.MustRegister(
			isOperationReadyAttempts,
			isOperationReadyWildcards,
			isOperationReadyDebugger,
			isOperationReadyDebuggerFailed,
			isOperationReadyDuration,
			isOperationDebugReadyDuration,
			durationPerGas,
			debugDurationPerGas,
			isOperationReadyError,
			isOperationReadyTrue,
			isOperationReadyFalse,
			isOperationReadyReverts,
			failoverSimulationAttempts,
			failoverSimulationSuccess,
			failoverSimulationError,
			failoverSimulationDuration,
			dependencyStateDuration,
			constraintsMetDuration,
			constraintsMet,
			constraintsNotMet,
			dependencyStateError,
			constraintsMetError,
			constraintMetDuration,
			dependencySlotDuration,
		)
	}

	return &metrics{
		isOperationReadyAttempts:       isOperationReadyAttempts,
		isOperationReadyWildcards:      isOperationReadyWildcards,
		isOperationReadyDebugger:       isOperationReadyDebugger,
		isOperationReadyDebuggerFailed: isOperationReadyDebuggerFailed,
		isOperationReadyDuration:       isOperationReadyDuration,
		isOperationDebugReadyDuration:  isOperationDebugReadyDuration,
		durationPerGas:                 durationPerGas,
		debugDurationPerGas:            debugDurationPerGas,
		isOperationReadyError:          isOperationReadyError,
		isOperationReadyTrue:           isOperationReadyTrue,
		isOperationReadyFalse:          isOperationReadyFalse,
		isOperationReadyReverts:        isOperationReadyReverts,
		failoverSimulationAttempts:     failoverSimulationAttempts,
		failoverSimulationSuccess:      failoverSimulationSuccess,
		failoverSimulationError:        failoverSimulationError,
		failoverSimulationDuration:     failoverSimulationDuration,
		dependencyStateDuration:        dependencyStateDuration,
		constraintsMetDuration:         constraintsMetDuration,
		constraintsMet:                 constraintsMet,
		constraintsNotMet:              constraintsNotMet,
		dependencyStateError:           dependencyStateError,
		constraintsMetError:            constraintsMetError,
		constraintMetDuration:          constraintMetDuration,
		dependencySlotDuration:         dependencySlotDuration,

		dependencyStateErrorBalance: prometheus.Labels{"reason": "balance"},
		dependencyStateErrorCode:    prometheus.Labels{"reason": "code"},
		dependencyStateErrorNonce:   prometheus.Labels{"reason": "nonce"},
		dependencyStateErrorSlots:   prometheus.Labels{"reason": "slots"},
	}
}
