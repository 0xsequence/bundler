package bundler

import (
	"context"
	"sync"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/lib/registry"
	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/go-chi/httplog/v2"
	"github.com/prometheus/client_golang/prometheus"
)

const PrunerBatchSize = 10
const PrunerWorkers = 3

type prunerMetrics struct {
	pruneBannedEmpty prometheus.Counter
	pruneBannedTime  prometheus.Histogram
	pruneBannedOps   prometheus.Counter

	pruneStaleAgeInf prometheus.Counter
	pruneStaleAge    prometheus.Histogram

	pruneStaleEmpty    prometheus.Counter
	pruneStaleTime     prometheus.Histogram
	pruneStaleReleased prometheus.Counter
	pruneStaleDropped  prometheus.Counter
	pruneStaleNoop     prometheus.Counter
	pruneStaleFailed   *prometheus.CounterVec

	failedPruneDependencyState prometheus.Labels
	failedPruneHasChanged      prometheus.Labels
}

func createPrunerMetrics(reg prometheus.Registerer) *prunerMetrics {
	pruneBannedTime := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "pruner_banned_time",
		Help:    "Time taken to prune banned operations",
		Buckets: prometheus.ExponentialBuckets(1e-6, 2, 15),
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
		Name: "pruner_stale_released_sum",
		Help: "Number of stale operations released",
	})

	pruneStaleDropped := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pruner_stale_dropped_sum",
		Help: "Number of stale operations dropped",
	})

	pruneStaleFailed := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pruner_stale_failed_sum",
		Help: "Number of failed stale operations",
	}, []string{"reason"})

	pruneBannedEmpty := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pruner_banned_empty",
		Help: "Number of empty banned runs",
	})

	pruneStaleEmpty := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pruner_stale_empty_sum",
		Help: "Number of empty stale runs",
	})

	pruneStaleNoop := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pruner_stale_noop_sum",
		Help: "Number of runs of operations that don't need to be re-evaluated",
	})

	pruneStaleAgeInf := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pruner_stale_age_inf",
		Help: "Stale operations that have readyAt set to zero",
	})

	pruneStaleAge := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "pruner_stale_age",
		Help:    "Age of stale operations that need to be re-evaluated",
		Buckets: prometheus.ExponentialBuckets(1, 2, 16),
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
			pruneStaleNoop,
			pruneStaleAge,
			pruneStaleAgeInf,
		)
	}

	return &prunerMetrics{
		pruneBannedEmpty:           pruneBannedEmpty,
		pruneBannedTime:            pruneBannedTime,
		pruneBannedOps:             pruneBannedOps,
		pruneStaleAgeInf:           pruneStaleAgeInf,
		pruneStaleAge:              pruneStaleAge,
		pruneStaleEmpty:            pruneStaleEmpty,
		pruneStaleTime:             pruneStaleTime,
		pruneStaleReleased:         pruneStaleReleased,
		pruneStaleDropped:          pruneStaleDropped,
		pruneStaleFailed:           pruneStaleFailed,
		pruneStaleNoop:             pruneStaleNoop,
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
	wg := sync.WaitGroup{}

	jobsChan := make(chan *mempool.TrackedOperation)
	for i := 0; i < PrunerWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.staleWorker(ctx, jobsChan)
		}()
	}

	if !s.NoStalePruning {
		wg.Add(1)
		go func() {
			s.staleFetcher(ctx, jobsChan)
		}()
	}

	if !s.NoBannedPruning {
		wg.Add(1)
		go func() {
			s.pruneBanned(ctx)
		}()

	}

	wg.Wait()
	close(jobsChan)
}

func (s *Pruner) pruneBanned(ctx context.Context) {
	for ctx.Err() == nil {
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

		// Wait for a bit, 10 seconds
		// but listen for context cancellation
		select {
		case <-time.After(10 * time.Second):
		case <-ctx.Done():
			return
		}
	}
}

func (s *Pruner) staleFetcher(ctx context.Context, jobsChan chan *mempool.TrackedOperation) {
	for ctx.Err() == nil {
		ops := s.Mempool.ReserveOps(ctx, func(to []*mempool.TrackedOperation) []*mempool.TrackedOperation {
			// Pick one operation above the grace period
			// start from the oldest part (the upper indexes)
			picked := make([]*mempool.TrackedOperation, 0, PrunerBatchSize)
			for i := len(to) - 1; i >= 0; i-- {
				if time.Since(to[i].ReadyAt) > s.GracePeriod {
					picked = append(picked, to[i])
					if len(picked) >= PrunerBatchSize {
						break
					}
				}
			}

			return picked
		})

		if len(ops) == 0 {
			s.metrics.pruneStaleEmpty.Inc()
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Attempt to feed the jobs channel
		for _, op := range ops {
			jobsChan <- op
		}
	}
}

func (s *Pruner) staleWorker(ctx context.Context, jobsChan chan *mempool.TrackedOperation) {
	for {
		select {
		case job := <-jobsChan:
			s.doStaleJob(ctx, job)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Pruner) doStaleJob(ctx context.Context, op *mempool.TrackedOperation) {
	start := time.Now()
	defer func() {
		s.metrics.pruneStaleTime.Observe(time.Since(start).Seconds())
	}()

	// Report how long has the operation been without being re-evaluated
	if op.ReadyAt.IsZero() {
		s.metrics.pruneStaleAgeInf.Inc()
	} else {
		s.metrics.pruneStaleAge.Observe(time.Since(op.ReadyAt).Seconds())
	}

	needsReevaluation := true
	if !op.EndorserResult.WildcardOnly {
		nextState, err := s.Endorser.DependencyState(ctx, op.EndorserResult)
		if err != nil {
			s.metrics.pruneStaleFailed.With(s.metrics.failedPruneDependencyState).Inc()
			s.logger.Error("pruner: error getting state", "error", err)
			// TODO: Handle error operations, ideally
			// we only allow an operation to fail a few times
			s.Mempool.DiscardOps(ctx, []string{op.Hash()})
			return
		}

		needsReevaluation, err = op.EndorserResult.HasChanged(op.EndorserResultState, nextState)
		if err != nil {
			s.metrics.pruneStaleFailed.With(s.metrics.failedPruneHasChanged).Inc()
			s.logger.Error("pruner: error comparing state", "error", err)
			// TODO: Handle error operations, ideally
			// we only allow an operation to fail a few times
			s.Mempool.DiscardOps(ctx, []string{op.Hash()})
			return
		}
	}

	if !needsReevaluation {
		// Release the operation
		s.metrics.pruneStaleNoop.Inc()
		s.Mempool.ReleaseOps(ctx, []string{op.Hash()}, proto.ReadyAtChange_Now)
		return
	}

	// We need to re-validate the operation
	res, err := s.Endorser.IsOperationReady(ctx, &op.Operation)
	if err != nil || !res.Readiness {
		s.metrics.pruneStaleDropped.Inc()
		// NOTICE This may not be an error, some endorsers revert instead of returning false
		s.Mempool.DiscardOps(ctx, []string{op.Hash()})
		return
	}

	s.metrics.pruneStaleReleased.Inc()
	s.Mempool.ReleaseOps(ctx, []string{op.Hash()}, proto.ReadyAtChange_Now)

}
