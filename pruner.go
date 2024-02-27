package bundler

import (
	"context"
	"time"

	"github.com/0xsequence/bundler/endorser"
	"github.com/go-chi/httplog/v2"
)

const PrunerBatchSize = 1

type Pruner struct {
	logger *httplog.Logger

	Mempool  MempoolInterface
	Endorser endorser.Interface
}

func NewPruner(mempool MempoolInterface, endorser endorser.Interface, logger *httplog.Logger) *Pruner {
	return &Pruner{
		logger: logger,

		Mempool:  mempool,
		Endorser: endorser,
	}
}

func (s *Pruner) Run(ctx context.Context) {
	for ctx.Err() == nil {
		ops := s.Mempool.ReserveOps(ctx, func(to []*TrackedOperation) []*TrackedOperation {
			var ops []*TrackedOperation

			// Pick the last `PrunerBatchSize` operations
			if PrunerBatchSize < len(to) {
				ops = to[len(to)-PrunerBatchSize:]
			} else {
				ops = to
			}

			oldops := make([]*TrackedOperation, 0, len(ops))
			for _, op := range ops {
				if time.Since(op.ReadyAt) > 5*time.Second {
					oldops = append(oldops, op)
				}
			}
			return oldops
		})

		failedOps := make([]*TrackedOperation, 0, len(ops))
		discartOps := make([]*TrackedOperation, 0, len(ops))
		releaseOps := make([]*TrackedOperation, 0, len(ops))

		// TODO: Batch this
		for _, op := range ops {
			nextState, err := s.Endorser.DependencyState(ctx, op.EndorserResult)
			if err != nil {
				s.logger.Error("pruner: error getting state", "error", err)
				failedOps = append(failedOps, op)
				continue
			}

			changed, err := op.EndorserResult.HasChanged(op.EndorserResultState, nextState)
			if err != nil {
				s.logger.Error("pruner: error comparing state", "error", err)
				failedOps = append(failedOps, op)
				continue
			}

			if changed {
				// We need to re-validate the operation
				// NOTICE that the endorser may revert instead of returning false
				res, err := s.Endorser.IsOperationReady(ctx, &op.Operation)
				if err != nil {
					discartOps = append(discartOps, op)
					continue
				}

				if !res.Readiness {
					discartOps = append(discartOps, op)
				} else {
					// TODO: handle the new set of dependencies
					releaseOps = append(releaseOps, op)
				}
			} else {
				// Release the operation
				releaseOps = append(releaseOps, op)
			}
		}

		// Release the operations

		if len(releaseOps) != 0 {
			s.logger.Debug("pruner: releasing operations", "operations", len(releaseOps))
			s.Mempool.ReleaseOps(ctx, releaseOps, ReadyAtChangeNow)
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

		// Sleep 1 second
		time.Sleep(time.Second)
	}
}
