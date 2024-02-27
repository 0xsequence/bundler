package mocks

import (
	"context"
	"time"

	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/types"
	"github.com/stretchr/testify/mock"
)

type MockMempool struct {
	mock.Mock
}

func (m *MockMempool) Size() int {
	return m.Called().Int(0)
}

func (m *MockMempool) IsKnownOp(op *types.Operation) bool {
	return m.Called(op).Bool(0)
}

func (m *MockMempool) AddOperation(ctx context.Context, op *types.Operation, forceInclude bool) error {
	return m.Called(ctx, op, forceInclude).Error(0)
}

func (m *MockMempool) ReserveOps(ctx context.Context, selectFn func([]*mempool.TrackedOperation) []*mempool.TrackedOperation) []*mempool.TrackedOperation {
	args := m.Called(ctx, selectFn)
	return args.Get(0).([]*mempool.TrackedOperation)
}

func (m *MockMempool) ReleaseOps(ctx context.Context, ops []*mempool.TrackedOperation, updateReadyAt mempool.ReadyAtChange) {
	m.Called(ctx, ops, updateReadyAt)
}

func (m *MockMempool) DiscardOps(ctx context.Context, ops []*mempool.TrackedOperation) {
	m.Called(ctx, ops)
}

func (m *MockMempool) ForgetOps(age time.Duration) []string {
	return m.Called(age).Get(0).([]string)
}

func (m *MockMempool) Inspect() *proto.MempoolView {
	return m.Called().Get(0).(*proto.MempoolView)
}

func (m *MockMempool) KnownOperations() []string {
	return m.Called().Get(0).([]string)
}

var _ mempool.Interface = &MockMempool{}