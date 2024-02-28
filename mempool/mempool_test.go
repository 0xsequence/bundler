package mempool_test

import (
	"testing"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/mocks"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/types"
	"github.com/go-chi/httplog/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

func TestAddOperation(t *testing.T) {
	logger := httplog.NewLogger("")
	mockP2p := &mocks.MockP2p{}
	mockCollector := &mocks.MockCollector{}
	mockEndorser := &mocks.MockEndorser{}

	mempool, err := mempool.NewMempool(&config.MempoolConfig{}, logger, mockEndorser, mockP2p, mockCollector, nil)

	assert.NoError(t, err)

	op := &types.Operation{}
	er := &endorser.EndorserResult{
		Readiness: true,
	}
	es := &endorser.EndorserResultState{}

	mockEndorser.On("IsOperationReady", mock.Anything, op).Return(er, nil).Once()
	mockEndorser.On("ConstraintsMet", mock.Anything, er).Return(true, nil).Once()
	mockEndorser.On("DependencyState", mock.Anything, er).Return(es, nil).Once()
	mockCollector.On("MeetsPayment", op).Return(true, nil).Once()

	mt := proto.MessageType_NEW_OPERATION
	mockP2p.On("Broadcast", proto.Message{
		Type:    &mt,
		Message: op.ToProto(),
	}).Return(nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	err = mempool.AddOperation(ctx, op, false)

	assert.NoError(t, err)
	cancel()

	mockEndorser.AssertExpectations(t)
	mockCollector.AssertExpectations(t)
	mockP2p.AssertExpectations(t)

	// The op should be known now
	assert.True(t, mempool.IsKnownOp(op))
	assert.Equal(t, len(mempool.Operations), 1)
	assert.Equal(t, mempool.Operations[0].ToProto(), op.ToProto())
}

func TestForceIncludeKnownOp(t *testing.T) {
	logger := httplog.NewLogger("")
	mockP2p := &mocks.MockP2p{}
	mockCollector := &mocks.MockCollector{}
	mockEndorser := &mocks.MockEndorser{}

	mempool, err := mempool.NewMempool(&config.MempoolConfig{}, logger, mockEndorser, mockP2p, mockCollector, nil)

	assert.NoError(t, err)

	op := &types.Operation{}
	er := &endorser.EndorserResult{
		Readiness: true,
	}
	es := &endorser.EndorserResultState{}

	mockEndorser.On("IsOperationReady", mock.Anything, op).Return(er, nil).Twice()
	mockEndorser.On("ConstraintsMet", mock.Anything, er).Return(true, nil).Twice()
	mockEndorser.On("DependencyState", mock.Anything, er).Return(es, nil).Twice()
	mockCollector.On("MeetsPayment", op).Return(true, nil).Twice()
	mockP2p.On("Broadcast", mock.Anything).Return(nil).Twice()

	ctx, cancel := context.WithCancel(context.Background())
	err = mempool.AddOperation(ctx, op, false)
	assert.NoError(t, err)

	// The op should be known now
	err = mempool.AddOperation(ctx, op, true)
	assert.NoError(t, err)

	mockEndorser.AssertExpectations(t)
	mockCollector.AssertExpectations(t)
	mockP2p.AssertExpectations(t)

	cancel()
}

func TestSkipAddingKnownOperation(t *testing.T) {
	logger := httplog.NewLogger("")
	mockP2p := &mocks.MockP2p{}
	mockCollector := &mocks.MockCollector{}
	mockEndorser := &mocks.MockEndorser{}

	mempool, err := mempool.NewMempool(&config.MempoolConfig{}, logger, mockEndorser, mockP2p, mockCollector, nil)

	assert.NoError(t, err)

	op := &types.Operation{}
	er := &endorser.EndorserResult{
		Readiness: true,
	}
	es := &endorser.EndorserResultState{}

	mockEndorser.On("IsOperationReady", mock.Anything, op).Return(er, nil).Once()
	mockEndorser.On("ConstraintsMet", mock.Anything, er).Return(true, nil).Once()
	mockEndorser.On("DependencyState", mock.Anything, er).Return(es, nil).Once()
	mockCollector.On("MeetsPayment", op).Return(true, nil).Once()
	mockP2p.On("Broadcast", mock.Anything).Return(nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	err = mempool.AddOperation(ctx, op, false)
	assert.NoError(t, err)

	// The op should be known now
	err = mempool.AddOperation(ctx, op, false)
	assert.Error(t, err)

	cancel()
}

func TestNotReadyOperation(t *testing.T) {
	logger := httplog.NewLogger("")
	mockP2p := &mocks.MockP2p{}
	mockCollector := &mocks.MockCollector{}
	mockEndorser := &mocks.MockEndorser{}

	mempool, err := mempool.NewMempool(&config.MempoolConfig{}, logger, mockEndorser, mockP2p, mockCollector, nil)

	assert.NoError(t, err)

	op := &types.Operation{}

	ctx, cancel := context.WithCancel(context.Background())

	mockEndorser.On("IsOperationReady", mock.Anything, op).Return(&endorser.EndorserResult{
		Readiness: false,
	}, nil).Once()

	err = mempool.AddOperation(ctx, op, false)
	assert.Error(t, err)
	f := mempool.ForgetOps(0)
	assert.Equal(t, f, []string{op.Digest()})

	// Maybe IsOperationReady returns an error
	mockEndorser.On("IsOperationReady", mock.Anything, op).Return(nil, assert.AnError).Once()

	err = mempool.AddOperation(ctx, op, false)
	assert.Error(t, err)
	f = mempool.ForgetOps(0)
	assert.Equal(t, f, []string{op.Digest()})

	mockEndorser.On("IsOperationReady", mock.Anything, op).Return(&endorser.EndorserResult{
		Readiness: true,
	}, nil).Maybe()

	// Maybe the contraints are not met
	mockEndorser.On("ConstraintsMet", mock.Anything, mock.Anything).Return(false, nil).Once()

	err = mempool.AddOperation(ctx, op, false)
	assert.Error(t, err)
	f = mempool.ForgetOps(0)
	assert.Equal(t, f, []string{op.Digest()})

	// Maybe the contraints failed
	mockEndorser.On("ConstraintsMet", mock.Anything, mock.Anything).Return(false, assert.AnError).Once()

	err = mempool.AddOperation(ctx, op, false)
	assert.Error(t, err)
	f = mempool.ForgetOps(0)
	assert.Equal(t, f, []string{op.Digest()})

	mockEndorser.On("ConstraintsMet", mock.Anything, mock.Anything).Return(true, nil).Maybe()

	// Maybe the dependency state fails
	mockEndorser.On("DependencyState", mock.Anything, mock.Anything).Return(nil, assert.AnError).Once()

	err = mempool.AddOperation(ctx, op, false)
	assert.Error(t, err)
	f = mempool.ForgetOps(0)
	assert.Equal(t, f, []string{op.Digest()})

	mockEndorser.On("DependencyState", mock.Anything, mock.Anything).Return(&endorser.EndorserResultState{}, nil).Maybe()

	// Maybe the collector fails
	mockCollector.On("MeetsPayment", op).Return(false, assert.AnError).Once()

	err = mempool.AddOperation(ctx, op, false)
	assert.Error(t, err)
	f = mempool.ForgetOps(0)
	assert.Equal(t, f, []string{op.Digest()})

	// Maybe the collector rejects it
	mockCollector.On("MeetsPayment", op).Return(false, nil).Once()

	err = mempool.AddOperation(ctx, op, false)
	assert.Error(t, err)
	f = mempool.ForgetOps(0)
	assert.Equal(t, f, []string{op.Digest()})

	cancel()
}

func TestReserveOps(t *testing.T) {
	logger := httplog.NewLogger("")
	mockP2p := &mocks.MockP2p{}
	mockCollector := &mocks.MockCollector{}
	mockEndorser := &mocks.MockEndorser{}

	mem, err := mempool.NewMempool(&config.MempoolConfig{}, logger, mockEndorser, mockP2p, mockCollector, nil)

	assert.NoError(t, err)

	op1 := &types.Operation{
		Calldata: []byte{0x01},
	}
	op2 := &types.Operation{
		Calldata: []byte{0x02},
	}
	op3 := &types.Operation{
		Calldata: []byte{0x03},
	}
	er := &endorser.EndorserResult{
		Readiness: true,
	}
	es := &endorser.EndorserResultState{}

	mockEndorser.On("IsOperationReady", mock.Anything, op1).Return(er, nil).Once()
	mockEndorser.On("IsOperationReady", mock.Anything, op2).Return(er, nil).Once()
	mockEndorser.On("IsOperationReady", mock.Anything, op3).Return(er, nil).Once()

	mockEndorser.On("ConstraintsMet", mock.Anything, er).Return(true, nil).Maybe()
	mockEndorser.On("DependencyState", mock.Anything, er).Return(es, nil).Maybe()
	mockCollector.On("MeetsPayment", mock.Anything).Return(true, nil).Maybe()
	mockP2p.On("Broadcast", mock.Anything).Return(nil).Maybe()

	ctx, cancel := context.WithCancel(context.Background())

	err = mem.AddOperation(ctx, op1, false)
	assert.NoError(t, err)
	err = mem.AddOperation(ctx, op2, false)
	assert.NoError(t, err)
	err = mem.AddOperation(ctx, op3, false)
	assert.NoError(t, err)

	// Reserve the first 2 ops
	reserved := mem.ReserveOps(ctx, func(to []*mempool.TrackedOperation) []*mempool.TrackedOperation {
		assert.Equal(t, len(to), 3)
		return to[:2]
	})

	assert.Equal(t, len(reserved), 2)
	assert.Equal(t, reserved[0].Operation.Digest(), op1.Digest())
	assert.Equal(t, reserved[1].Operation.Digest(), op2.Digest())

	// Calling reserve again should only give one option
	reserved = mem.ReserveOps(ctx, func(to []*mempool.TrackedOperation) []*mempool.TrackedOperation {
		assert.Equal(t, len(to), 1)
		return []*mempool.TrackedOperation{}
	})

	cancel()
}
