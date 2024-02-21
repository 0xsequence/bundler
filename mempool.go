package bundler

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/p2p"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/types"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/go-chi/httplog/v2"
)

type ReadyAtChange int

const (
	ReadyAtChangeNone ReadyAtChange = iota
	ReadyAtChangeNow
	ReadyAtChangeZero
)

type TrackedOperation struct {
	types.Operation

	ReservedSince *time.Time `json:"reserved_since,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	ReadyAt       time.Time  `json:"ready_at"`

	EndorserResult      *endorser.EndorserResult      `json:"endorser_result,omitempty"`
	EndorserResultState *endorser.EndorserResultState `json:"endorser_result_state,omitempty"`
}

type Mempool struct {
	logger  *httplog.Logger
	ipfsurl string

	Host     *p2p.Host
	Provider *ethrpc.Provider
	MaxSize  uint

	flock           sync.Mutex
	FreshOperations *[]*types.Operation

	olock      sync.Mutex
	Operations []*TrackedOperation

	digests map[string]struct{}
}

func NewMempool(cfg *config.MempoolConfig, logger *httplog.Logger, provider *ethrpc.Provider, host *p2p.Host) (*Mempool, error) {
	mp := &Mempool{
		logger:  logger,
		ipfsurl: cfg.IpfsUrl,

		Host:     host,
		Provider: provider,
		MaxSize:  cfg.Size,

		flock: sync.Mutex{},
		olock: sync.Mutex{},

		FreshOperations: &[]*types.Operation{},
		Operations:      []*TrackedOperation{},

		digests: map[string]struct{}{},
	}

	return mp, nil
}

func (mp *Mempool) Size() int {
	return len(mp.Operations) + len(*mp.FreshOperations)
}

func (mp *Mempool) mustBeUniqueOp(op *types.Operation) error {
	digest := op.Digest()
	if _, ok := mp.digests[digest]; ok {
		return fmt.Errorf("mempool: duplicate operation")
	}

	mp.digests[digest] = struct{}{}
	return nil
}

func (mp *Mempool) AddOperationSync(ctx context.Context, op *types.Operation) error {
	if op == nil {
		return fmt.Errorf("mempool: operation is nil")
	}

	err := mp.mustBeUniqueOp(op)
	if err != nil {
		return err
	}

	// NOTICE: Adding operations in sync does not respect the max size
	return mp.tryPromoteOperation(ctx, op)
}

func (mp *Mempool) AddOperation(op *types.Operation) error {
	if op == nil {
		return fmt.Errorf("mempool: operation is nil")
	}

	err := mp.mustBeUniqueOp(op)
	if err != nil {
		return err
	}

	mp.flock.Lock()
	defer mp.flock.Unlock()

	if mp.Size() >= int(mp.MaxSize) {
		return fmt.Errorf("mempool: max size reached")
	}

	mp.logger.Info("mempool: adding operation to fresh", "op", op.Digest())

	nlist := append(*mp.FreshOperations, op)
	mp.FreshOperations = &nlist

	return nil
}

func (mp *Mempool) ReserveOps(ctx context.Context, selectFn func([]*TrackedOperation) []*TrackedOperation) []*TrackedOperation {
	mp.olock.Lock()
	defer mp.olock.Unlock()

	// Filter out the operations that are already reserved
	// and the ones that are not ready
	availOps := []*TrackedOperation{}
	for _, op := range mp.Operations {
		if op.ReservedSince != nil {
			continue
		}
		availOps = append(availOps, op)
	}

	// Select the operations to reserve
	resOps := selectFn(availOps)
	for _, op := range resOps {
		n := time.Now()
		op.ReservedSince = &n
	}

	return resOps
}

func (mp *Mempool) ReleaseOps(ctx context.Context, ops []*TrackedOperation, updateReadyAt ReadyAtChange) {
	mp.olock.Lock()
	defer mp.olock.Unlock()

	for _, op := range mp.Operations {
		for _, rop := range ops {
			if op.Digest() == rop.Digest() {
				rop.ReservedSince = nil

				switch updateReadyAt {
				case ReadyAtChangeNow:
					rop.ReadyAt = time.Now()
				case ReadyAtChangeZero:
					rop.ReadyAt = time.Time{}
				}
			}
		}
	}

	mp.SortOperations()
}

func (mp *Mempool) SortOperations() {
	sort.Slice(mp.Operations, func(i, j int) bool {
		return mp.Operations[i].ReadyAt.Before(mp.Operations[j].ReadyAt)
	})
}

func (mp *Mempool) DiscardOps(ctx context.Context, ops []*TrackedOperation) {
	mp.olock.Lock()
	defer mp.olock.Unlock()

	var kops []*TrackedOperation
	for _, op := range mp.Operations {
		discard := false

		for _, dop := range ops {
			if op.Digest() == dop.Digest() {
				discard = true
				break
			}
		}

		if discard {
			continue
		}

		kops = append(kops, op)
	}

	mp.Operations = kops
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
	mp.FreshOperations = &[]*types.Operation{}
	mp.flock.Unlock()

	for _, op := range *freshOps {
		err := mp.tryPromoteOperation(ctx, op)
		if err != nil {
			mp.logger.Warn("error adding operation to mempool", "op", op.Digest(), "err", err)
		}
	}

	mp.olock.Lock()
	mp.SortOperations()
	mp.olock.Unlock()

	return nil
}

func (mp *Mempool) tryPromoteOperation(ctx context.Context, op *types.Operation) error {
	res, err := endorser.IsOperationReady(ctx, mp.Provider, op)
	if err != nil {
		return fmt.Errorf("IsOperationReady failed: %w", err)
	}

	if !res.Readiness {
		return fmt.Errorf("operation not ready")
	}

	// Check the constraints
	okc, err := endorser.CheckDependencyConstraints(ctx, res.Dependencies, mp.Provider)
	if err != nil {
		return fmt.Errorf("CheckDependencyConstraints failed: %w", err)
	}

	if !okc {
		return fmt.Errorf("operation constraints not met")
	}

	state, err := res.State(ctx, mp.Provider)
	if err != nil {
		return fmt.Errorf("EndorserResultState failed: %w", err)
	}

	// If the operation is ready
	// then we add it to the mempool

	mp.logger.Info("operation added to mempool", "op", op.Digest())
	mp.ReportToIPFS(op)

	mp.olock.Lock()
	defer mp.olock.Unlock()

	mp.Operations = append(mp.Operations, &TrackedOperation{
		Operation: *op,

		CreatedAt: time.Now(),
		ReadyAt:   time.Now(),

		EndorserResult:      res,
		EndorserResultState: state,
	})

	// Broadcast the operation to the network
	// ONLY now, since we are sure it's ready
	if mp.Host != nil {
		fmt.Println("Broadcasting operation to the network")
		messageType := proto.MessageType_NEW_OPERATION
		err = mp.Host.Broadcast(proto.Message{
			Type:    &messageType,
			Message: op.ToProto(),
		})
		if err != nil {
			mp.logger.Warn("error broadcasting operation to the network", "op", op.Digest(), "err", err)
		}
	}

	return nil
}

func (mp *Mempool) ReportToIPFS(op *types.Operation) {
	// Fire a go-routine to report the operation to IPFS
	if mp.ipfsurl == "" {
		return
	}

	go func() {
		err := op.ReportToIPFS(mp.ipfsurl)
		if err != nil {
			mp.logger.Warn("error reporting operation to IPFS", "op", op.Digest(), "err", err)
		}
	}()
}

func (mp *Mempool) Inspect(ctx context.Context) *proto.MempoolView {
	mp.olock.Lock()
	defer mp.olock.Unlock()

	lockCount := 0
	ops := make([]*TrackedOperation, 0, len(mp.Operations))
	for i := range mp.Operations {
		ops = append(ops, mp.Operations[i])
		if mp.Operations[i].ReservedSince != nil {
			lockCount++
		}
	}

	fops := make([]*types.Operation, 0, len(*mp.FreshOperations))
	fops = append(fops, *mp.FreshOperations...)

	seen := make([]string, 0, len(mp.digests))
	for k := range mp.digests {
		seen = append(seen, k)
	}

	return &proto.MempoolView{
		FreshOperationsSize: len(*mp.FreshOperations),
		OperationsSize:      len(mp.Operations),
		SeenSize:            len(mp.digests),
		LockSize:            lockCount,

		Seen:            seen,
		Operations:      ops,
		FreshOperations: fops,
	}
}
