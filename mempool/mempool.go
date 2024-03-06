package mempool

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/0xsequence/bundler/collector"
	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/ipfs"
	"github.com/0xsequence/bundler/mempool/partitioner"
	"github.com/0xsequence/bundler/p2p"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/types"
	"github.com/go-chi/httplog/v2"
)

type Mempool struct {
	logger *httplog.Logger

	Ipfs      ipfs.Interface
	Host      p2p.Interface
	Collector collector.Interface
	Endorser  endorser.Interface

	MaxSize int

	lock       sync.Mutex
	Operations []*TrackedOperation

	partitioner *partitioner.Partitioner
	known       *KnownOperations
}

var _ Interface = &Mempool{}

func NewMempool(cfg *config.MempoolConfig, logger *httplog.Logger, endorser endorser.Interface, host p2p.Interface, collector collector.Interface, ipfs ipfs.Interface) (*Mempool, error) {
	if cfg.Size <= 1 {
		return nil, fmt.Errorf("mempool: size must be greater than 1")
	}

	overLapLimit := cfg.OverlapLimit
	if overLapLimit <= 0 {
		logger.Warn("mempool: overlap limit is less than 1, setting to 1")
		overLapLimit = 1
	}

	wildcardLimit := cfg.WildcardLimit
	if wildcardLimit <= 0 {
		logger.Warn("mempool: wildcard limit is less than 1, setting to 1")
		wildcardLimit = 1
	}

	mp := &Mempool{
		logger: logger,

		Ipfs:      ipfs,
		Host:      host,
		Endorser:  endorser,
		Collector: collector,

		MaxSize: int(cfg.Size),

		Operations: []*TrackedOperation{},

		partitioner: partitioner.NewPartitioner(overLapLimit, wildcardLimit),

		known: &KnownOperations{
			lock:    sync.RWMutex{},
			digests: map[string]time.Time{},
		},
	}

	return mp, nil
}

func (mp *Mempool) Size() int {
	return len(mp.Operations)
}

func (mp *Mempool) IsKnownOp(op *types.Operation) bool {
	mp.known.lock.RLock()
	defer mp.known.lock.RUnlock()

	_, ok := mp.known.digests[op.Hash()]
	return ok
}

func (mp *Mempool) mustBeUniqueOp(op *types.Operation) error {
	digest := op.Hash()

	mp.known.lock.Lock()
	defer mp.known.lock.Unlock()

	if _, ok := mp.known.digests[digest]; ok {
		return fmt.Errorf("mempool: duplicate operation")
	}

	// Time zero means that it was not marked
	// for removal.
	mp.known.digests[digest] = time.Time{}

	return nil
}

func (mp *Mempool) AddOperation(ctx context.Context, op *types.Operation, forceInclude bool) error {
	if op == nil {
		return fmt.Errorf("mempool: operation is nil")
	}

	// We save the op but we don't fail
	// if it already exists
	err := mp.mustBeUniqueOp(op)
	if err != nil && !forceInclude {
		return err
	}

	// NOTICE: Adding operations in sync does not respect the max size
	err = mp.tryPromoteOperation(ctx, op)

	// If it fails, we need to mark the operation
	// for deletion, or else it will hang around forever
	if err != nil {
		mp.markForForget(op)
	}

	return err
}

func (mp *Mempool) ReserveOps(ctx context.Context, selectFn func([]*TrackedOperation) []*TrackedOperation) []*TrackedOperation {
	mp.lock.Lock()
	defer mp.lock.Unlock()

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

func (mp *Mempool) ReleaseOps(ctx context.Context, ops []string, updateReadyAt proto.ReadyAtChange) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	for _, op := range mp.Operations {
		for _, rop := range ops {
			if op.Hash() == rop {
				op.ReservedSince = nil

				switch updateReadyAt {
				case proto.ReadyAtChange_Now:
					op.ReadyAt = time.Now()
				case proto.ReadyAtChange_Zero:
					op.ReadyAt = time.Time{}
				}
			}
		}
	}

	mp.SortOperations()
}

func (mp *Mempool) SortOperations() {
	sort.Slice(mp.Operations, func(i, j int) bool {
		return mp.Operations[i].ReadyAt.After(mp.Operations[j].ReadyAt)
	})
}

func (mp *Mempool) DiscardOps(ctx context.Context, ops []string) {
	mp.lock.Lock()
	defer mp.lock.Unlock()
	mp.discardOpsUnlocked(ctx, ops)
}

func (mp *Mempool) discardOpsUnlocked(ctx context.Context, ops []string) {
	var kops []*TrackedOperation
	for _, op := range mp.Operations {
		discard := false

		for _, dop := range ops {
			if op.Hash() == dop {
				discard = true
				break
			}
		}

		if discard {
			// Mark the operation for deletion by setting
			// the time to the current time
			mp.markForForget(&op.Operation)
			continue
		}

		kops = append(kops, op)
	}

	mp.Operations = kops
	mp.partitioner.Remove(ops)
}

func (mp *Mempool) markForForget(op *types.Operation) {
	mp.known.lock.Lock()
	defer mp.known.lock.Unlock()

	mp.known.digests[op.Hash()] = time.Now()
}

func (mp *Mempool) evictLesser(ctx context.Context, cand *types.Operation, subsets *[][]*types.Operation) error {
	var groups [][]*types.Operation
	if subsets == nil {
		// Having a standalone method for evicting lesser operations
		// on the whole mempool *could* be faster, but we keep it simple for now
		groups = make([][]*types.Operation, 1)
		groups[0] = make([]*types.Operation, len(mp.Operations))
		for i, op := range mp.Operations {
			groups[0][i] = &op.Operation
		}
	} else {
		groups = *subsets
	}

	evictions := make([]string, 0, len(groups))
	for _, alts := range groups {
		worst := cand
		var secondWorstVal *big.Int = nil

		for _, alt := range alts {
			av := alt.Value()
			if av.Cmp(worst.Value()) < 0 {
				secondWorstVal = worst.Value()
				worst = alt
			} else if secondWorstVal == nil || av.Cmp(secondWorstVal) < 0 {
				secondWorstVal = alt.Value()
			}
		}

		// Don't evict if the worst is the candidate
		if worst == cand {
			return fmt.Errorf("candidate is the worst operation: %s - %s < %s", worst.Hash(), worst.Value(), secondWorstVal)
		}

		evictions = append(evictions, worst.Hash())
	}

	mp.discardOpsUnlocked(ctx, evictions)

	return nil
}

func (mp *Mempool) tryPromoteOperation(ctx context.Context, op *types.Operation) error {
	res, err := mp.Endorser.IsOperationReady(ctx, op)
	if err != nil {
		return fmt.Errorf("IsOperationReady failed: %w", err)
	}

	if !res.Readiness {
		return fmt.Errorf("operation not ready")
	}

	// Check the constraints
	okc, err := mp.Endorser.ConstraintsMet(ctx, res)
	if err != nil {
		return fmt.Errorf("CheckDependencyConstraints failed: %w", err)
	}

	if !okc {
		return fmt.Errorf("operation constraints not met")
	}

	state, err := mp.Endorser.DependencyState(ctx, res)
	if err != nil {
		return fmt.Errorf("EndorserResultState failed: %w", err)
	}

	// Check the collector (fees)
	if err := mp.Collector.ValidatePayment(op); err != nil {
		return fmt.Errorf("payment validation failed: %w", err)
	}

	// If the operation is ready
	// then we add it to the mempool

	mp.logger.Info("operation added to mempool", "op", op.Hash())
	go mp.ReportToIPFS(op)

	mp.lock.Lock()
	defer mp.lock.Unlock()

	// Check the dependency overlap
	ok, overlaps := mp.partitioner.Add(op, res)
	if !ok {
		// Evict lesser operations among the overlaps and see if we can
		// make room for the new operation
		err := mp.evictLesser(ctx, op, &overlaps)
		if err != nil {
			return err
		}

		// Try again, now we should have room
		// notice that a race condition could happen here
		// but in that case we fail
		ok, _ = mp.partitioner.Add(op, res)
		if !ok {
			return fmt.Errorf("operation dependency constraints not met")
		}
	} else if len(mp.Operations) >= mp.MaxSize {
		// We need to evict *something* to make room
		err := mp.evictLesser(ctx, op, nil)
		if err != nil {
			return err
		}
	}

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
		err = mp.Host.Broadcast(proto.Message{
			Type:    proto.MessageType_NEW_OPERATION,
			Message: op.ToProto(),
		})
		if err != nil {
			mp.logger.Warn("error broadcasting operation to the network", "op", op.Hash(), "err", err)
		}
	}

	return nil
}

func (mp *Mempool) ReportToIPFS(op *types.Operation) {
	// Fire a go-routine to report the operation to IPFS
	if mp.Ipfs == nil {
		return
	}

	go func() {
		err := op.ReportToIPFS(mp.Ipfs)
		if err != nil {
			mp.logger.Warn("error reporting operation to IPFS", "op", op.Hash(), "err", err)
		}
	}()
}

func (mp *Mempool) ForgetOps(age time.Duration) []string {
	mp.known.lock.Lock()
	defer mp.known.lock.Unlock()

	forgotten := make([]string, 0, len(mp.known.digests))
	nt := time.Time{}

	for k, v := range mp.known.digests {
		if v != nt && time.Since(v) > age {
			forgotten = append(forgotten, k)
			delete(mp.known.digests, k)
		}
	}

	return forgotten
}

func (mp *Mempool) KnownOperations() []string {
	mp.known.lock.RLock()
	defer mp.known.lock.RUnlock()

	ops := make([]string, 0, len(mp.known.digests))
	for k := range mp.known.digests {
		ops = append(ops, k)
	}

	return ops
}

func (mp *Mempool) Inspect() *proto.MempoolView {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	lockCount := 0
	ops := make([]*TrackedOperation, 0, len(mp.Operations))
	for i := range mp.Operations {
		ops = append(ops, mp.Operations[i])
		if mp.Operations[i].ReservedSince != nil {
			lockCount++
		}
	}

	mp.known.lock.Lock()
	defer mp.known.lock.Unlock()

	seen := make([]string, 0, len(mp.known.digests))
	for k := range mp.known.digests {
		seen = append(seen, k)
	}

	return &proto.MempoolView{
		Size:     len(mp.Operations),
		SeenSize: len(mp.known.digests),
		LockSize: lockCount,

		Seen:       seen,
		Operations: ops,
	}
}
