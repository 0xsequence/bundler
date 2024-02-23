package bundler

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/ipfs"
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

type KnownOperations struct {
	lock    sync.RWMutex
	digests map[string]time.Time
}

type Mempool struct {
	logger *httplog.Logger
	ipfs   *ipfs.Client

	Host     *p2p.Host
	Provider *ethrpc.Provider
	MaxSize  uint

	lock       sync.Mutex
	Operations []*TrackedOperation

	known *KnownOperations
}

func NewMempool(cfg *config.MempoolConfig, logger *httplog.Logger, provider *ethrpc.Provider, host *p2p.Host, ipfs *ipfs.Client) (*Mempool, error) {
	mp := &Mempool{
		logger: logger,
		ipfs:   ipfs,

		Host:     host,
		Provider: provider,
		MaxSize:  cfg.Size,

		Operations: []*TrackedOperation{},

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

	_, ok := mp.known.digests[op.Digest()]
	return ok
}

func (mp *Mempool) mustBeUniqueOp(op *types.Operation) error {
	digest := op.Digest()

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
	return mp.tryPromoteOperation(ctx, op)
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

func (mp *Mempool) ReleaseOps(ctx context.Context, ops []*TrackedOperation, updateReadyAt ReadyAtChange) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

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
	mp.lock.Lock()
	defer mp.lock.Unlock()

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

		// Mark the operation for deletion by setting
		// the time to the current time
		digest := op.Digest()
		mp.known.lock.Lock()
		mp.known.digests[digest] = time.Now()
		mp.known.lock.Unlock()

		kops = append(kops, op)
	}

	mp.Operations = kops
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

	mp.lock.Lock()
	defer mp.lock.Unlock()

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
	if mp.ipfs == nil {
		return
	}

	go func() {
		err := op.ReportToIPFS(mp.ipfs)
		if err != nil {
			mp.logger.Warn("error reporting operation to IPFS", "op", op.Digest(), "err", err)
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

func (mp *Mempool) Inspect(ctx context.Context) *proto.MempoolView {
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
