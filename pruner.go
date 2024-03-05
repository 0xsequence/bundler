package bundler

import (
	"context"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/proto"
	"github.com/go-chi/httplog/v2"
)

const PrunerBatchSize = 1

type Pruner struct {
	GracePeriod time.Duration
	RunWait     time.Duration
	logger      *httplog.Logger

	Mempool  mempool.Interface
	Endorser endorser.Interface
}

func NewPruner(cfg config.PrunerConfig, mempool mempool.Interface, endorser endorser.Interface, logger *httplog.Logger) *Pruner {
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
		GracePeriod: gracePeriod,
		RunWait:     runWait,

		logger: logger,

		Mempool:  mempool,
		Endorser: endorser,
	}
}

func (s *Pruner) Run(ctx context.Context) {
	for ctx.Err() == nil {
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

		failedOps := make([]string, 0, len(ops))
		discartOps := make([]string, 0, len(ops))
		releaseOps := make([]string, 0, len(ops))

		// TODO: Batch this
		for _, op := range ops {
			nextState, err := s.Endorser.DependencyState(ctx, op.EndorserResult)
			if err != nil {
				s.logger.Error("pruner: error getting state", "error", err)
				failedOps = append(failedOps, op.Hash())
				continue
			}

			changed, err := op.EndorserResult.HasChanged(op.EndorserResultState, nextState)
			if err != nil {
				s.logger.Error("pruner: error comparing state", "error", err)
				failedOps = append(failedOps, op.Hash())
				continue
			}

			if changed {
				// We need to re-validate the operation
				// NOTICE that the endorser may revert instead of returning false
				res, err := s.Endorser.IsOperationReady(ctx, &op.Operation)
				if err != nil {
					discartOps = append(discartOps, op.Hash())
					continue
				}

				if !res.Readiness {
					discartOps = append(discartOps, op.Hash())
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
			s.logger.Debug("pruner: releasing operations", "operations", len(releaseOps))
			s.Mempool.ReleaseOps(ctx, releaseOps, proto.ReadyAtChange_Now)
		}

		if len(discartOps) != 0 {
			s.logger.Info("pruner: discarding operations", "operations", len(discartOps))
			s.Mempool.DiscardOps(ctx, discartOps)
		}

		// TODO: Handle error operations, ideally
		// we only allow an operation to fail a few times

		if len(failedOps) != 0 {
			s.logger.Warn("pruner: failed operations", "operations", len(failedOps))
			s.Mempool.DiscardOps(ctx, failedOps)
		}

		time.Sleep(s.RunWait)
	}
}
