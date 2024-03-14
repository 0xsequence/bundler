package mempool_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/0xsequence/bundler/calldata"
	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/contracts/gen/solabis/abiendorser"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/mocks"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/types"
	"github.com/0xsequence/ethkit/go-ethereum/common"
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
	mockRegistry := &mocks.MockRegistry{}

	mempool, err := mempool.NewMempool(&config.MempoolConfig{
		Size: 10,
	}, logger, mockEndorser, mockP2p, mockCollector, nil, calldata.DefaultModel(), mockRegistry)

	assert.NoError(t, err)

	op := &types.Operation{}
	er := &endorser.EndorserResult{
		Readiness: true,
	}
	es := &endorser.EndorserResultState{}

	mockEndorser.On("IsOperationReady", mock.Anything, op).Return(er, nil).Once()
	mockEndorser.On("ConstraintsMet", mock.Anything, er).Return(true, nil).Once()
	mockEndorser.On("DependencyState", mock.Anything, er).Return(es, nil).Once()
	mockCollector.On("ValidatePayment", op).Return(nil).Once()
	mockRegistry.On("IsAcceptedEndorser", common.Address{}).Return(true).Once()

	mockP2p.On("Broadcast", proto.Message{
		Type:    proto.MessageType_NEW_OPERATION,
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

	mockEndorser.AssertExpectations(t)
	mockCollector.AssertExpectations(t)
	mockP2p.AssertExpectations(t)
	mockRegistry.AssertExpectations(t)
}

func TestForceIncludeKnownOp(t *testing.T) {
	logger := httplog.NewLogger("")
	mockP2p := &mocks.MockP2p{}
	mockCollector := &mocks.MockCollector{}
	mockEndorser := &mocks.MockEndorser{}
	mockRegistry := &mocks.MockRegistry{}

	mempool, err := mempool.NewMempool(&config.MempoolConfig{
		Size: 10,
	}, logger, mockEndorser, mockP2p, mockCollector, nil, calldata.DefaultModel(), mockRegistry)

	assert.NoError(t, err)

	op := &types.Operation{}
	er := &endorser.EndorserResult{
		Readiness: true,
	}
	es := &endorser.EndorserResultState{}

	mockEndorser.On("IsOperationReady", mock.Anything, op).Return(er, nil).Twice()
	mockEndorser.On("ConstraintsMet", mock.Anything, er).Return(true, nil).Twice()
	mockEndorser.On("DependencyState", mock.Anything, er).Return(es, nil).Twice()
	mockCollector.On("ValidatePayment", op).Return(nil).Twice()
	mockP2p.On("Broadcast", mock.Anything).Return(nil).Twice()
	mockRegistry.On("IsAcceptedEndorser", common.Address{}).Return(true).Twice()

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

	mockEndorser.AssertExpectations(t)
	mockCollector.AssertExpectations(t)
	mockP2p.AssertExpectations(t)
	mockRegistry.AssertExpectations(t)
}

func TestSkipAddingKnownOperation(t *testing.T) {
	logger := httplog.NewLogger("")
	mockP2p := &mocks.MockP2p{}
	mockCollector := &mocks.MockCollector{}
	mockEndorser := &mocks.MockEndorser{}
	mockRegistry := &mocks.MockRegistry{}

	mempool, err := mempool.NewMempool(&config.MempoolConfig{
		Size: 10,
	}, logger, mockEndorser, mockP2p, mockCollector, nil, calldata.DefaultModel(), mockRegistry)

	assert.NoError(t, err)

	op := &types.Operation{}
	er := &endorser.EndorserResult{
		Readiness: true,
	}
	es := &endorser.EndorserResultState{}

	mockEndorser.On("IsOperationReady", mock.Anything, op).Return(er, nil).Once()
	mockEndorser.On("ConstraintsMet", mock.Anything, er).Return(true, nil).Once()
	mockEndorser.On("DependencyState", mock.Anything, er).Return(es, nil).Once()
	mockCollector.On("ValidatePayment", op).Return(nil).Once()
	mockP2p.On("Broadcast", mock.Anything).Return(nil).Once()
	mockRegistry.On("IsAcceptedEndorser", common.Address{}).Return(true).Once()

	ctx, cancel := context.WithCancel(context.Background())
	err = mempool.AddOperation(ctx, op, false)
	assert.NoError(t, err)

	// The op should be known now
	err = mempool.AddOperation(ctx, op, false)
	assert.Error(t, err)

	cancel()

	mockEndorser.AssertExpectations(t)
	mockCollector.AssertExpectations(t)
	mockP2p.AssertExpectations(t)
	mockRegistry.AssertExpectations(t)
}

func TestNotReadyOperation(t *testing.T) {
	logger := httplog.NewLogger("")
	mockP2p := &mocks.MockP2p{}
	mockCollector := &mocks.MockCollector{}
	mockEndorser := &mocks.MockEndorser{}
	mockRegistry := &mocks.MockRegistry{}

	mempool, err := mempool.NewMempool(&config.MempoolConfig{
		Size: 10,
	}, logger, mockEndorser, mockP2p, mockCollector, nil, calldata.DefaultModel(), mockRegistry)

	assert.NoError(t, err)

	op := &types.Operation{}

	ctx, cancel := context.WithCancel(context.Background())

	mockEndorser.On("IsOperationReady", mock.Anything, op).Return(&endorser.EndorserResult{
		Readiness: false,
	}, nil).Once()

	err = mempool.AddOperation(ctx, op, false)
	assert.Error(t, err)
	f := mempool.ForgetOps(0)
	assert.Equal(t, f, []string{op.Hash()})

	// Maybe IsOperationReady returns an error
	mockEndorser.On("IsOperationReady", mock.Anything, op).Return(nil, assert.AnError).Once()

	err = mempool.AddOperation(ctx, op, false)
	assert.Error(t, err)
	f = mempool.ForgetOps(0)
	assert.Equal(t, f, []string{op.Hash()})

	mockEndorser.On("IsOperationReady", mock.Anything, op).Return(&endorser.EndorserResult{
		Readiness: true,
	}, nil).Maybe()

	// Maybe the constraints are not met
	mockEndorser.On("ConstraintsMet", mock.Anything, mock.Anything).Return(false, nil).Once()

	err = mempool.AddOperation(ctx, op, false)
	assert.Error(t, err)
	f = mempool.ForgetOps(0)
	assert.Equal(t, f, []string{op.Hash()})

	// Maybe the constraints failed
	mockEndorser.On("ConstraintsMet", mock.Anything, mock.Anything).Return(false, assert.AnError).Once()

	err = mempool.AddOperation(ctx, op, false)
	assert.Error(t, err)
	f = mempool.ForgetOps(0)
	assert.Equal(t, f, []string{op.Hash()})

	mockEndorser.On("ConstraintsMet", mock.Anything, mock.Anything).Return(true, nil).Maybe()

	// Maybe the dependency state fails
	mockEndorser.On("DependencyState", mock.Anything, mock.Anything).Return(nil, assert.AnError).Once()

	err = mempool.AddOperation(ctx, op, false)
	assert.Error(t, err)
	f = mempool.ForgetOps(0)
	assert.Equal(t, f, []string{op.Hash()})

	mockEndorser.On("DependencyState", mock.Anything, mock.Anything).Return(&endorser.EndorserResultState{}, nil).Maybe()

	// Maybe the collector fails
	mockCollector.On("ValidatePayment", op).Return(assert.AnError).Once()

	err = mempool.AddOperation(ctx, op, false)
	assert.Error(t, err)
	f = mempool.ForgetOps(0)
	assert.Equal(t, f, []string{op.Hash()})

	cancel()
}

func TestReserveOps(t *testing.T) {
	logger := httplog.NewLogger("")
	mockP2p := &mocks.MockP2p{}
	mockCollector := &mocks.MockCollector{}
	mockEndorser := &mocks.MockEndorser{}
	mockRegistry := &mocks.MockRegistry{}

	mem, err := mempool.NewMempool(&config.MempoolConfig{
		Size: 10,
	}, logger, mockEndorser, mockP2p, mockCollector, nil, calldata.DefaultModel(), mockRegistry)

	assert.NoError(t, err)

	op1 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{0x01},
		},
	}
	op2 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{0x02},
		},
	}
	op3 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{0x03},
		},
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
	mockCollector.On("ValidatePayment", mock.Anything).Return(nil).Maybe()
	mockP2p.On("Broadcast", mock.Anything).Return(nil).Maybe()
	mockRegistry.On("IsAcceptedEndorser", mock.Anything).Return(true).Maybe()

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
	assert.Equal(t, reserved[0].Operation.Hash(), op1.Hash())
	assert.Equal(t, reserved[1].Operation.Hash(), op2.Hash())

	// Calling reserve again should only give one option
	reserved2 := mem.ReserveOps(ctx, func(to []*mempool.TrackedOperation) []*mempool.TrackedOperation {
		assert.Equal(t, len(to), 1)
		return []*mempool.TrackedOperation{}
	})

	assert.Equal(t, len(reserved2), 0)

	// Release the reserved ops
	// sleep a bit so op2 readyAt is newer than op3, op1 goes to zero
	time.Sleep(10 * time.Millisecond)
	mem.ReleaseOps(ctx, []string{reserved[0].Hash()}, proto.ReadyAtChange_Zero)
	mem.ReleaseOps(ctx, []string{reserved[1].Hash()}, proto.ReadyAtChange_Now)

	// Should sort the operations from the most recent ready at first
	reserved = mem.ReserveOps(ctx, func(to []*mempool.TrackedOperation) []*mempool.TrackedOperation {
		return to
	})

	assert.Equal(t, len(reserved), 3)

	// The new order should be: op2, op3, op1
	assert.Equal(t, reserved[0].Operation.Hash(), op2.Hash())
	assert.Equal(t, reserved[1].Operation.Hash(), op3.Hash())
	assert.Equal(t, reserved[2].Operation.Hash(), op1.Hash())

	// Discard only two operations
	mem.DiscardOps(ctx, []string{reserved[0].Hash(), reserved[1].Hash()})
	mem.ReleaseOps(ctx, []string{reserved[2].Hash()}, proto.ReadyAtChange_Zero)

	// Reserving now should only give the last operation
	mem.ReserveOps(ctx, func(to []*mempool.TrackedOperation) []*mempool.TrackedOperation {
		assert.Equal(t, len(to), 1)
		assert.Equal(t, to[0].Operation.Hash(), op1.Hash())
		return []*mempool.TrackedOperation{}
	})

	// They now should be marked for forget
	f := mem.ForgetOps(0)
	assert.Contains(t, f, op2.Hash())
	assert.Contains(t, f, op3.Hash())

	cancel()

	mockEndorser.AssertExpectations(t)
}

func TestReportToIPFS(t *testing.T) {
	logger := httplog.NewLogger("")
	mockP2p := &mocks.MockP2p{}
	mockCollector := &mocks.MockCollector{}
	mockEndorser := &mocks.MockEndorser{}
	mockIpfs := &mocks.MockIpfs{}
	mockRegistry := &mocks.MockRegistry{}

	mempool, err := mempool.NewMempool(&config.MempoolConfig{
		Size: 10,
	}, logger, mockEndorser, mockP2p, mockCollector, mockIpfs, calldata.DefaultModel(), mockRegistry)

	assert.NoError(t, err)

	op1 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{0x01},
		},
	}

	// Should report to IPFS if the operation is valid
	mockEndorser.On("ConstraintsMet", mock.Anything, mock.Anything).Return(true, nil).Maybe()
	mockEndorser.On("DependencyState", mock.Anything, mock.Anything).Return(&endorser.EndorserResultState{}, nil).Maybe()
	mockCollector.On("ValidatePayment", op1).Return(nil).Maybe()
	mockP2p.On("Broadcast", mock.Anything).Return(nil).Maybe()
	mockRegistry.On("IsAcceptedEndorser", mock.Anything).Return(true).Maybe()

	mockEndorser.On("IsOperationReady", mock.Anything, op1).Return(&endorser.EndorserResult{
		Readiness: true,
	}, nil).Once()

	done := make(chan struct{})

	mockIpfs.On("Report", mock.Anything).Run(func(mock.Arguments) {
		done <- struct{}{}
	}).Return(op1.Hash(), nil).Once()

	ctx, cancel := context.WithCancel(context.Background())

	err = mempool.AddOperation(ctx, op1, false)
	assert.NoError(t, err)

	<-done

	mockIpfs.AssertExpectations(t)

	// Do not report to IPFS if it fails
	op2 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{0x02},
		},
	}

	mockEndorser.On("IsOperationReady", mock.Anything, op2).Return(&endorser.EndorserResult{
		Readiness: false,
	}, nil).Once()

	err = mempool.AddOperation(ctx, op2, false)
	assert.Error(t, err)

	mockIpfs.AssertExpectations(t)

	cancel()
}

func TestRejectOverlappingDependency(t *testing.T) {
	logger := httplog.NewLogger("")
	mockP2p := &mocks.MockP2p{}
	mockCollector := &mocks.MockCollector{}
	mockEndorser := &mocks.MockEndorser{}
	mockRegistry := &mocks.MockRegistry{}

	mempool, err := mempool.NewMempool(&config.MempoolConfig{
		Size:         10,
		OverlapLimit: 1,
	}, logger, mockEndorser, mockP2p, mockCollector, nil, calldata.DefaultModel(), mockRegistry)

	assert.NoError(t, err)

	op1 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data:                   []byte{0x01},
			GasLimit:               big.NewInt(2),
			MaxFeePerGas:           big.NewInt(1),
			FeeScalingFactor:       big.NewInt(1),
			FeeNormalizationFactor: big.NewInt(1),
		},
	}

	op2 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data:                   []byte{0x02},
			GasLimit:               big.NewInt(1),
			MaxFeePerGas:           big.NewInt(1),
			FeeScalingFactor:       big.NewInt(1),
			FeeNormalizationFactor: big.NewInt(1),
		},
	}

	mockEndorser.On("ConstraintsMet", mock.Anything, mock.Anything).Return(true, nil).Maybe()
	mockEndorser.On("DependencyState", mock.Anything, mock.Anything).Return(&endorser.EndorserResultState{}, nil).Maybe()
	mockCollector.On("ValidatePayment", mock.Anything).Return(nil).Maybe()
	mockCollector.On("Cmp", mock.Anything, mock.Anything).Return(0).Maybe()
	mockP2p.On("Broadcast", mock.Anything).Return(nil).Maybe()
	mockRegistry.On("IsAcceptedEndorser", mock.Anything).Return(true).Maybe()

	mockEndorser.On("IsOperationReady", mock.Anything, mock.Anything).Return(&endorser.EndorserResult{
		Readiness: true,
		Dependencies: []endorser.Dependency{{
			Addr:    common.HexToAddress("0x5887Ea54AE1308Bb7A697FdE87bA3D2E2d3952Ad"),
			Balance: true,
		}},
	}, nil).Twice()

	res := mempool.AddOperation(context.Background(), op1, false)
	assert.NoError(t, res)

	res = mempool.AddOperation(context.Background(), op2, false)
	assert.Error(t, res)
}

func TestReplaceOverlappingDependency(t *testing.T) {
	logger := httplog.NewLogger("")
	mockP2p := &mocks.MockP2p{}
	mockCollector := &mocks.MockCollector{}
	mockEndorser := &mocks.MockEndorser{}
	mockRegistry := &mocks.MockRegistry{}

	mempool, err := mempool.NewMempool(&config.MempoolConfig{
		Size:         10,
		OverlapLimit: 1,
	}, logger, mockEndorser, mockP2p, mockCollector, nil, calldata.DefaultModel(), mockRegistry)

	assert.NoError(t, err)

	op1 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data:                   []byte{0x01},
			GasLimit:               big.NewInt(1),
			MaxFeePerGas:           big.NewInt(1),
			MaxPriorityFeePerGas:   big.NewInt(1),
			FeeScalingFactor:       big.NewInt(1),
			FeeNormalizationFactor: big.NewInt(1),
		},
	}

	op2 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data:                   []byte{0x02},
			GasLimit:               big.NewInt(2),
			MaxFeePerGas:           big.NewInt(1),
			MaxPriorityFeePerGas:   big.NewInt(1),
			FeeScalingFactor:       big.NewInt(1),
			FeeNormalizationFactor: big.NewInt(1),
		},
	}

	mockEndorser.On("ConstraintsMet", mock.Anything, mock.Anything).Return(true, nil).Maybe()
	mockEndorser.On("DependencyState", mock.Anything, mock.Anything).Return(&endorser.EndorserResultState{}, nil).Maybe()
	mockCollector.On("ValidatePayment", mock.Anything).Return(nil).Maybe()
	mockCollector.On("Cmp", op1, op2).Return(-1).Maybe()
	mockCollector.On("Cmp", op2, op1).Return(1).Maybe()
	mockP2p.On("Broadcast", mock.Anything).Return(nil).Maybe()
	mockRegistry.On("IsAcceptedEndorser", mock.Anything).Return(true).Maybe()

	mockEndorser.On("IsOperationReady", mock.Anything, mock.Anything).Return(&endorser.EndorserResult{
		Readiness: true,
		Dependencies: []endorser.Dependency{{
			Addr:    common.HexToAddress("0x5887Ea54AE1308Bb7A697FdE87bA3D2E2d3952Ad"),
			Balance: true,
		}},
	}, nil).Twice()

	res := mempool.AddOperation(context.Background(), op1, false)
	assert.NoError(t, res)

	res = mempool.AddOperation(context.Background(), op2, false)
	assert.NoError(t, res)

	assert.Equal(t, len(mempool.Operations), 1)
	assert.Equal(t, mempool.Operations[0].ToProto(), op2.ToProto())
}

func TestRejectBadEndorser(t *testing.T) {
	logger := httplog.NewLogger("")
	mockP2p := &mocks.MockP2p{}
	mockCollector := &mocks.MockCollector{}
	mockEndorser := &mocks.MockEndorser{}
	mockRegistry := &mocks.MockRegistry{}

	mempool, err := mempool.NewMempool(&config.MempoolConfig{
		Size:          10,
		OverlapLimit:  10,
		WildcardLimit: 10,
	}, logger, mockEndorser, mockP2p, mockCollector, nil, calldata.DefaultModel(), mockRegistry)

	assert.NoError(t, err)

	op1 := &types.Operation{
		Endorser: common.HexToAddress("0x5887Ea54AE1308Bb7A697FdE87bA3D2E2d3952Ad"),
	}

	mockEndorser.On("IsOperationReady", mock.Anything, op1).Return(&endorser.EndorserResult{
		Readiness: true,
	}, nil).Once()
	mockEndorser.On("ConstraintsMet", mock.Anything, mock.Anything).Return(true, nil).Maybe()
	mockEndorser.On("DependencyState", mock.Anything, mock.Anything).Return(&endorser.EndorserResultState{}, nil).Maybe()
	mockCollector.On("ValidatePayment", mock.Anything).Return(nil).Maybe()
	mockP2p.On("Broadcast", mock.Anything).Return(nil).Maybe()
	mockRegistry.On("IsAcceptedEndorser", common.HexToAddress("0x5887Ea54AE1308Bb7A697FdE87bA3D2E2d3952Ad")).Return(false).Once()

	ctx, cancel := context.WithCancel(context.Background())
	res := mempool.AddOperation(ctx, op1, false)
	assert.Error(t, res)

	cancel()

	mockEndorser.AssertExpectations(t)
	mockCollector.AssertExpectations(t)
	mockP2p.AssertExpectations(t)
	mockRegistry.AssertExpectations(t)
}
