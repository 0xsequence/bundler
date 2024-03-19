package bundler

import (
	"context"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/registry"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/go-chi/httplog/v2"
	"github.com/prometheus/client_golang/prometheus"
)

const PrunerBatchSize = 1

type prunerMetrics struct {
	pruneBannedEmpty prometheus.Counter
	pruneBannedTime  prometheus.Histogram
	pruneBannedOps   prometheus.Counter

	pruneStaleEmpty    prometheus.Counter
	pruneStaleTime     prometheus.Histogram
	pruneStaleReleased prometheus.Counter
	pruneStaleDropped  prometheus.Counter
	pruneStaleFailed   *prometheus.CounterVec

	failedPruneDependencyState prometheus.Labels
	failedPruneHasChanged      prometheus.Labels
}

func createPrunerMetrics(reg prometheus.Registerer) *prunerMetrics {
	pruneBannedTime := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "pruner_banned_time",
		Help: "Time taken to prune banned operations",
		Buckets: []float64{
			1e-6, // 0.000001 seconds
			2e-6, // 0.000002 seconds
			3e-6, // 0.000003 seconds
			4e-6, // 0.000004 seconds
			5e-6, // 0.000005 seconds
			1e-5, // 0.00001 seconds
			2e-5, // 0.00002 seconds
			5e-5, // 0.00005 seconds
			1e-4, // 0.0001 seconds
			2e-4, // 0.0002 seconds
			5e-4, // 0.0005 seconds
			1e-3, // 0.001 seconds (1 millisecond)
			1e-2, // 0.01 seconds (10 milliseconds)
		},
	})

	pruneBannedOps := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pruner_banned_ops",
		Help: "Number of banned operations pruned",
	})

	pruneStaleTime := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "pruner_stale_time",
		Help:    "Time taken to prune stale operations",
		Buckets: prometheus.DefBuckets,
	})

	pruneStaleReleased := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pruner_stale_released",
		Help: "Number of stale operations released",
	})

	pruneStaleDropped := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pruner_stale_dropped",
		Help: "Number of stale operations dropped",
	})

	pruneStaleFailed := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pruner_stale_failed",
		Help: "Number of failed stale operations",
	}, []string{"reason"})

	pruneBannedEmpty := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pruner_banned_empty",
		Help: "Number of empty banned runs",
	})

	pruneStaleEmpty := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pruner_stale_empty",
		Help: "Number of empty stale runs",
	})

	failedPruneDependencyState := prometheus.Labels{
		"reason": "dependency_state",
	}

	failedPruneHasChanged := prometheus.Labels{
		"reason": "has_changed",
	}

	if reg != nil {
		reg.MustRegister(
			pruneBannedTime,
			pruneBannedOps,
			pruneStaleTime,
			pruneStaleReleased,
			pruneStaleDropped,
			pruneStaleFailed,
			pruneBannedEmpty,
			pruneStaleEmpty,
		)
	}

	return &prunerMetrics{
		pruneBannedEmpty:           pruneBannedEmpty,
		pruneBannedTime:            pruneBannedTime,
		pruneBannedOps:             pruneBannedOps,
		pruneStaleEmpty:            pruneStaleEmpty,
		pruneStaleTime:             pruneStaleTime,
		pruneStaleReleased:         pruneStaleReleased,
		pruneStaleDropped:          pruneStaleDropped,
		pruneStaleFailed:           pruneStaleFailed,
		failedPruneDependencyState: failedPruneDependencyState,
		failedPruneHasChanged:      failedPruneHasChanged,
	}
}

type Pruner struct {
	NoStalePruning  bool
	NoBannedPruning bool

	GracePeriod time.Duration
	RunWait     time.Duration

	logger  *httplog.Logger
	metrics *prunerMetrics

	Mempool  mempool.Interface
	Endorser endorser.Interface
	Registry registry.Interface
}

func NewPruner(cfg config.PrunerConfig, logger *httplog.Logger, metrics prometheus.Registerer, mempool mempool.Interface, endorser endorser.Interface, registry registry.Interface) *Pruner {
	var gracePeriod time.Duration
	if cfg.GracePeriodSeconds == 0 {
		gracePeriod = 5 * time.Second
	} else {
		gracePeriod = time.Duration(cfg.GracePeriodSeconds) * time.Second
	}

	var runWait time.Duration
	if cfg.RunWaitMillis == 0 {
		runWait = 1 * time.Second
	} else {
		runWait = time.Duration(cfg.RunWaitMillis) * time.Millisecond
	}

	if logger != nil {
		logger.Info("pruner: grace period", "seconds", gracePeriod.Seconds())
		logger.Info("pruner: run wait", "milliseconds", runWait.Milliseconds())
	}

	return &Pruner{
		NoStalePruning:  cfg.NoStalePruning,
		NoBannedPruning: cfg.NoBannedPruning,

		GracePeriod: gracePeriod,
		RunWait:     runWait,

		logger:  logger,
		metrics: createPrunerMetrics(metrics),

		Mempool:  mempool,
		Endorser: endorser,
		Registry: registry,
	}
}

func (s *Pruner) Run(ctx context.Context) {
	for ctx.Err() == nil {
		if !s.NoStalePruning {
			s.pruneStale(ctx)
		}

		if !s.NoBannedPruning {
			s.pruneBanned(ctx)
		}

		time.Sleep(s.RunWait)
	}
}

func (s *Pruner) pruneBanned(ctx context.Context) {
	// We don't want to hang the mempool, so we first get all the banned endorsers
	start := time.Now()

	allEndorsers := s.Registry.KnownEndorsers()
	bannedEndorsers := make(map[common.Address]struct{}, len(allEndorsers))
	for _, endorser := range allEndorsers {
		if endorser.Status == registry.PermanentBanned || endorser.Status == registry.TemporaryBanned {
			bannedEndorsers[endorser.Address] = struct{}{}
		}
	}

	ops := s.Mempool.ReserveOps(ctx, func(to []*mempool.TrackedOperation) []*mempool.TrackedOperation {
		var ops []*mempool.TrackedOperation

		for _, op := range to {
			if _, banned := bannedEndorsers[op.Endorser]; banned {
				ops = append(ops, op)
			}
		}

		return ops
	})

	// Discard the operations
	if len(ops) != 0 {
		opHashes := make([]string, len(ops))
		for i, op := range ops {
			opHashes[i] = op.Hash()
		}

		s.metrics.pruneBannedOps.Add(float64(len(ops)))
		s.logger.Info("pruner: discarding banned operations", "operations", len(ops))
		s.Mempool.DiscardOps(ctx, opHashes)
	} else {
		s.metrics.pruneBannedEmpty.Inc()
	}

	s.metrics.pruneBannedTime.Observe(time.Since(start).Seconds())
}

func (s *Pruner) pruneStale(ctx context.Context) {
	start := time.Now()

	ops := s.Mempool.ReserveOps(ctx, func(to []*mempool.TrackedOperation) []*mempool.TrackedOperation {
		var ops []*mempool.TrackedOperation

		// Pick the last `PrunerBatchSize` operations
		if PrunerBatchSize < len(to) {
			ops = to[len(to)-PrunerBatchSize:]
		} else {
			ops = to
		}

		oldops := make([]*mempool.TrackedOperation, 0, len(ops))
		for _, op := range ops {
			if time.Since(op.ReadyAt) > s.GracePeriod {
				oldops = append(oldops, op)
			}
		}
		return oldops
	})

	if len(ops) == 0 {
		s.metrics.pruneStaleEmpty.Inc()
		return
	}

	failedOps := make([]string, 0, len(ops))
	DiscardOps := make([]string, 0, len(ops))
	releaseOps := make([]string, 0, len(ops))

	// TODO: Batch this
	for _, op := range ops {
		var needsReevaluation bool

		if op.EndorserResult.WildcardOnly {
			// Wildcard operations always require validation, as we can't
			// validate the dependencies of them
			needsReevaluation = true
		} else {
			nextState, err := s.Endorser.DependencyState(ctx, op.EndorserResult)
			if err != nil {
				s.metrics.pruneStaleFailed.With(s.metrics.failedPruneDependencyState).Inc()
				s.logger.Error("pruner: error getting state", "error", err)
				failedOps = append(failedOps, op.Hash())
				continue
			}

			needsReevaluation, err = op.EndorserResult.HasChanged(op.EndorserResultState, nextState)
			if err != nil {
				s.metrics.pruneStaleFailed.With(s.metrics.failedPruneHasChanged).Inc()
				s.logger.Error("pruner: error comparing state", "error", err)
				failedOps = append(failedOps, op.Hash())
				continue
			}
		}

		if needsReevaluation {
			// We need to re-validate the operation
			// NOTICE that the endorser may revert instead of returning false
			res, err := s.Endorser.IsOperationReady(ctx, &op.Operation)
			if err != nil {
				DiscardOps = append(DiscardOps, op.Hash())
				continue
			}

			if !res.Readiness {
				DiscardOps = append(DiscardOps, op.Hash())
			} else {
				// TODO: handle the new set of dependencies
				releaseOps = append(releaseOps, op.Hash())
			}
		} else {
			// Release the operation
			releaseOps = append(releaseOps, op.Hash())
		}
	}

	// Release the operations

	if len(releaseOps) != 0 {
		s.metrics.pruneStaleReleased.Add(float64(len(releaseOps)))
		s.logger.Debug("pruner: releasing operations", "operations", len(releaseOps))
		s.Mempool.ReleaseOps(ctx, releaseOps, proto.ReadyAtChange_Now)
	}

	if len(DiscardOps) != 0 {
		s.metrics.pruneStaleDropped.Add(float64(len(DiscardOps)))
		s.logger.Info("pruner: discarding operations", "operations", len(DiscardOps))
		s.Mempool.DiscardOps(ctx, DiscardOps)
	}

	// TODO: Handle error operations, ideally
	// we only allow an operation to fail a few times

	if len(failedOps) != 0 {
		// Metric handled in the loop
		s.logger.Warn("pruner: failed operations", "operations", len(failedOps))
		s.Mempool.DiscardOps(ctx, failedOps)
	}

	s.metrics.pruneStaleTime.Observe(time.Since(start).Seconds())
}
