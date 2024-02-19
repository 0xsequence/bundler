package bundler

import (
	"context"
	"time"

	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/go-chi/httplog/v2"
)

const PrunerBatchSize = 1

type Pruner struct {
	logger *httplog.Logger

	Mempool  *Mempool
	Provider *ethrpc.Provider
}

func NewPruner(mempool *Mempool, provider *ethrpc.Provider, logger *httplog.Logger) *Pruner {
	return &Pruner{
		logger: logger,

		Mempool:  mempool,
		Provider: provider,
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
			nextState, err := op.EndorserResult.State(ctx, s.Provider)
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
				res, err := endorser.IsOperationReady(ctx, s.Provider, &op.Operation)
				if err != nil {
					s.logger.Error("pruner: error validating operation", "error", err)
					failedOps = append(failedOps, op)
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

		if len(releaseOps) != 0 {
			s.logger.Info("pruner: releasing operations", "operations", len(releaseOps))
		}

		if len(discartOps) != 0 {
			s.logger.Info("pruner: discarding operations", "operations", len(discartOps))
		}

		if len(failedOps) != 0 {
			s.logger.Warn("pruner: failed operations", "operations", len(failedOps))
		}

		// Release the operations
		s.Mempool.ReleaseOps(ctx, releaseOps, true)
		s.Mempool.DiscardOps(ctx, discartOps)

		// TODO: Handle error operations, ideally
		// we only allow an operation to fail a few times
		s.Mempool.DiscardOps(ctx, failedOps)

		// Sleep 1 second
		time.Sleep(time.Second)
	}
}
