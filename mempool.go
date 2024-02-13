package bundler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/go-chi/httplog/v2"
)

type TrackedOperation struct {
	proto.Operation

	EndorserResult *endorser.EndorserResult
}

type Mempool struct {
	logger *httplog.Logger

	Provider *ethrpc.Provider
	MaxSize  uint

	flock           sync.Mutex
	FreshOperations *[]*proto.Operation

	olock      sync.Mutex
	Operations []TrackedOperation
}

func NewMempool(cfg *config.MempoolConfig, logger *httplog.Logger, provider *ethrpc.Provider) (*Mempool, error) {
	mp := &Mempool{
		logger:   logger,
		Provider: provider,
		MaxSize:  cfg.Size,

		flock: sync.Mutex{},
		olock: sync.Mutex{},

		FreshOperations: &[]*proto.Operation{},
		Operations:      []TrackedOperation{},
	}

	return mp, nil
}

func (mp *Mempool) Size() int {
	return len(mp.Operations) + len(*mp.FreshOperations)
}

func (mp *Mempool) AddOperation(op *proto.Operation) error {
	if op == nil {
		return fmt.Errorf("mempool: operation is nil")
	}

	mp.flock.Lock()
	defer mp.flock.Unlock()

	if mp.Size() >= int(mp.MaxSize) {
		return fmt.Errorf("mempool: max size reached")
	}

	mp.logger.Info("mempool: adding operation to fresh", "op", op)

	nlist := append(*mp.FreshOperations, op)
	mp.FreshOperations = &nlist

	return nil
}

func (mp *Mempool) StartProcessor(ctx context.Context) {
	// Run every 500 ms
	go func() {
		ticker := time.NewTicker(time.Duration(500 * time.Millisecond))
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				err := mp.HandleFreshOps(ctx)
				if err != nil {
					mp.logger.Error("mempool: error handling fresh operations", "err", err)
				}
			case <-ctx.Done():
				// Context cancelled, stop the ticker
				return
			}
		}
	}()
}

func (mp *Mempool) HandleFreshOps(ctx context.Context) error {
	// Take each operation from the fresh mempool, pass them through the
	// endorser. If they are not ready, then we drop them.

	// TODO: Parallelize this

	if len(*mp.FreshOperations) == 0 {
		return nil
	}

	// Create a local copy of the fresh operations
	// and clear the fresh operations list. This is going
	// to take a while and we don't want to block new operations.

	// NOTICE that the mempool could grow over the limit while we are
	// processing the fresh operations. This is fine for now.

	mp.flock.Lock()
	freshOps := mp.FreshOperations
	mp.FreshOperations = &[]*proto.Operation{}
	mp.flock.Unlock()

	for _, op := range *freshOps {
		res, err := endorser.IsOperationReady(ctx, mp.Provider, op)
		if err != nil {
			mp.logger.Warn("dropping operation", "op", op, "reason", "endorser error", "err", err)
			continue
		}

		if !res.Readiness {
			mp.logger.Debug("dropping operation", "op", op, "reason", "not ready")
			continue
		}

		// If the operation is ready
		// then we add it to the mempool

		mp.olock.Lock()
		mp.logger.Info("operation added to mempool", "op", op)
		mp.Operations = append(mp.Operations, TrackedOperation{
			Operation:      *op,
			EndorserResult: res,
		})
		mp.olock.Unlock()
	}

	return nil
}
