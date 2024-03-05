package mempool

import (
	"context"
	"sync"
	"time"

	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/types"
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

type Interface interface {
	Size() int
	IsKnownOp(op *types.Operation) bool
	AddOperation(ctx context.Context, op *types.Operation, forceInclude bool) error
	ReserveOps(ctx context.Context, selectFn func([]*TrackedOperation) []*TrackedOperation) []*TrackedOperation
	ReleaseOps(ctx context.Context, ops []string, updateReadyAt proto.ReadyAtChange)
	DiscardOps(ctx context.Context, ops []string)
	ForgetOps(age time.Duration) []string
	KnownOperations() []string
	Inspect() *proto.MempoolView
}
