package bundler_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/0xsequence/bundler"
	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/contracts/gen/solabis/abiendorser"
	"github.com/0xsequence/bundler/mocks"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/types"
	"github.com/0xsequence/go-sequence/lib/prototyp"
	"github.com/go-chi/httplog/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIgnoreKnown(t *testing.T) {
	mockMempool := &mocks.MockMempool{}
	logger := httplog.NewLogger("")

	ingress := bundler.NewIngress(
		&config.MempoolConfig{},
		logger,
		nil,
		mockMempool,
		nil,
		nil,
	)

	op := &types.Operation{}

	mockMempool.On("IsKnownOp", op).Return(true).Once()

	err := ingress.Add(op)
	assert.NoError(t, err)
}

func TestRejectLowPayment(t *testing.T) {
	mockMempool := &mocks.MockMempool{}
	mockCollector := &mocks.MockCollector{}
	logger := httplog.NewLogger("")

	ingress := bundler.NewIngress(
		&config.MempoolConfig{},
		logger,
		nil,
		mockMempool,
		mockCollector,
		nil,
	)

	op := &types.Operation{}

	mockMempool.On("IsKnownOp", op).Return(false).Twice()
	mockCollector.On("ValidatePayment", op).Return(fmt.Errorf("err")).Once()
	err := ingress.Add(op)
	assert.Error(t, err)
}

func TestIgnoreInTransit(t *testing.T) {
	mockMempool := &mocks.MockMempool{}
	mockCollector := &mocks.MockCollector{}
	logger := httplog.NewLogger("")

	ingress := bundler.NewIngress(
		&config.MempoolConfig{
			IngressSize: 2,
		},
		logger,
		nil,
		mockMempool,
		mockCollector,
		nil,
	)

	op1 := &types.Operation{}
	op2 := &types.Operation{}

	mockMempool.On("IsKnownOp", op1).Return(false).Once()
	mockMempool.On("IsKnownOp", op2).Return(false).Once()
	mockCollector.On("ValidatePayment", op1).Return(nil).Once()
	mockCollector.On("ValidatePayment", op2).Return(nil).Once()

	err := ingress.Add(op1)
	assert.NoError(t, err)
	assert.Equal(t, 1, ingress.InBuffer())

	err = ingress.Add(op2)
	assert.NoError(t, err)
	assert.Equal(t, 1, ingress.InBuffer())
}

func TestRejectBufferFull(t *testing.T) {
	mockMempool := &mocks.MockMempool{}
	mockCollector := &mocks.MockCollector{}
	logger := httplog.NewLogger("")

	ingress := bundler.NewIngress(
		&config.MempoolConfig{
			IngressSize: 1,
		},
		logger,
		nil,
		mockMempool,
		mockCollector,
		nil,
	)

	op1 := &types.Operation{}
	op2 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			GasLimit: big.NewInt(1),
		},
	}

	mockMempool.On("IsKnownOp", op1).Return(false).Once()
	mockMempool.On("IsKnownOp", op2).Return(false).Once()
	mockCollector.On("ValidatePayment", op1).Return(nil).Once()
	mockCollector.On("ValidatePayment", op2).Return(nil).Once()

	err := ingress.Add(op1)
	assert.NoError(t, err)

	err = ingress.Add(op2)
	assert.Error(t, err)
}

func TestAddOperation(t *testing.T) {
	mockMempool := &mocks.MockMempool{}
	mockCollector := &mocks.MockCollector{}
	mockP2p := &mocks.MockP2p{}
	logger := httplog.NewLogger("")

	ingress := bundler.NewIngress(
		&config.MempoolConfig{
			IngressSize: 1,
		},
		logger,
		nil,
		mockMempool,
		mockCollector,
		mockP2p,
	)

	op := &types.Operation{}

	done := make(chan bool)

	mockP2p.On("HandleMessageType", proto.MessageType_NEW_OPERATION, mock.Anything).Return(nil).Once()

	mockMempool.On("IsKnownOp", op).Return(false).Once()
	mockMempool.On("AddOperation", mock.Anything, op, false).Run(func(args mock.Arguments) {
		done <- true

	}).Return(nil).Once()
	mockCollector.On("ValidatePayment", op).Return(nil).Once()

	go ingress.Run(context.Background())

	err := ingress.Add(op)
	assert.NoError(t, err)

	<-done

	assert.Equal(t, 0, ingress.InBuffer())
}

func TestBuffer(t *testing.T) {
	mockMempool := &mocks.MockMempool{}
	mockCollector := &mocks.MockCollector{}
	mockP2p := &mocks.MockP2p{}
	logger := httplog.NewLogger("")

	ingress := bundler.NewIngress(
		&config.MempoolConfig{
			IngressSize: 5,
		},
		logger,
		nil,
		mockMempool,
		mockCollector,
		mockP2p,
	)

	op1 := &types.Operation{}
	op2 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			GasLimit: big.NewInt(1),
		},
	}
	op3 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			GasLimit: big.NewInt(2),
		},
	}

	done := make(chan bool)

	mockMempool.On("IsKnownOp", mock.Anything).Return(false).Times(3)
	mockMempool.On("AddOperation", mock.Anything, op1, false).Run(func(args mock.Arguments) {
		done <- true
	}).Return(nil).Once()
	mockMempool.On("AddOperation", mock.Anything, op2, false).Run(func(args mock.Arguments) {
		done <- true
	}).Return(nil).Once()
	mockMempool.On("AddOperation", mock.Anything, op3, false).Run(func(args mock.Arguments) {
		done <- true
	}).Return(nil).Once()
	mockCollector.On("ValidatePayment", mock.Anything).Return(nil).Times(3)

	mockP2p.On("HandleMessageType", proto.MessageType_NEW_OPERATION, mock.Anything).Return(nil).Once()

	go ingress.Run(context.Background())

	err := ingress.Add(op1)
	assert.NoError(t, err)

	err = ingress.Add(op2)
	assert.NoError(t, err)

	err = ingress.Add(op3)
	assert.NoError(t, err)

	<-done
	<-done
	<-done

	assert.Equal(t, 0, ingress.InBuffer())
}

func TestListenP2P(t *testing.T) {
	mockMempool := &mocks.MockMempool{}
	mockCollector := &mocks.MockCollector{}
	mockP2p := &mocks.MockP2p{}
	logger := httplog.NewLogger("")

	ingress := bundler.NewIngress(
		&config.MempoolConfig{
			IngressSize: 5,
		},
		logger,
		nil,
		mockMempool,
		mockCollector,
		mockP2p,
	)

	op1 := &proto.Operation{
		Entrypoint:             "0x2430d0F4D8cF4A5594d953c9aF1Ed3E6772e3a83",
		Data:                   "0xc0decafe",
		GasLimit:               prototyp.BigInt(*big.NewInt(2)),
		FeeToken:               "0x2430d0F4D8cF4A5594d953c9aF1Ed3E6772e3a83",
		Endorser:               "0x2430d0F4D8cF4A5594d953c9aF1Ed3E6772e3a83",
		EndorserCallData:       "0x",
		EndorserGasLimit:       prototyp.BigInt(*big.NewInt(1)),
		MaxFeePerGas:           prototyp.BigInt(*big.NewInt(1)),
		MaxPriorityFeePerGas:   prototyp.BigInt(*big.NewInt(1)),
		FeeScalingFactor:       prototyp.BigInt(*big.NewInt(1)),
		FeeNormalizationFactor: prototyp.BigInt(*big.NewInt(1)),
		HasUntrustedContext:    false,
	}

	top1, err := types.NewOperationFromProto(op1)
	assert.NoError(t, err)

	op2 := &proto.Operation{
		Entrypoint:             "0x2430d0F4D8cF4A5594d953c9aF1Ed3E6772e3a83",
		Data:                   "0xc0decafe",
		GasLimit:               prototyp.BigInt(*big.NewInt(2)),
		FeeToken:               "0x2430d0F4D8cF4A5594d953c9aF1Ed3E6772e3a83",
		Endorser:               "0x2430d0F4D8cF4A5594d953c9aF1Ed3E6772e3a83",
		EndorserCallData:       "0x",
		EndorserGasLimit:       prototyp.BigInt(*big.NewInt(1)),
		MaxFeePerGas:           prototyp.BigInt(*big.NewInt(1)),
		MaxPriorityFeePerGas:   prototyp.BigInt(*big.NewInt(1)),
		FeeScalingFactor:       prototyp.BigInt(*big.NewInt(1)),
		FeeNormalizationFactor: prototyp.BigInt(*big.NewInt(2)),
		HasUntrustedContext:    false,
	}

	top2, err := types.NewOperationFromProto(op2)
	assert.NoError(t, err)

	op3 := &proto.Operation{
		Entrypoint:             "0x2430d0F4D8cF4A5594d953c9aF1Ed3E6772e3a83",
		Data:                   "0xc0decafe",
		GasLimit:               prototyp.BigInt(*big.NewInt(2)),
		FeeToken:               "0x2430d0F4D8cF4A5594d953c9aF1Ed3E6772e3a83",
		Endorser:               "0x2430d0F4D8cF4A5594d953c9aF1Ed3E6772e3a83",
		EndorserCallData:       "0x",
		EndorserGasLimit:       prototyp.BigInt(*big.NewInt(1)),
		MaxFeePerGas:           prototyp.BigInt(*big.NewInt(1)),
		MaxPriorityFeePerGas:   prototyp.BigInt(*big.NewInt(1)),
		FixedGas:               prototyp.BigInt(*big.NewInt(1)),
		FeeNormalizationFactor: prototyp.BigInt(*big.NewInt(3)),
		HasUntrustedContext:    false,
	}

	top3, err := types.NewOperationFromProto(op3)
	assert.NoError(t, err)

	done := make(chan bool)

	mockP2p.On("HandleMessageType", proto.MessageType_NEW_OPERATION, mock.Anything).Return(nil).Once()

	go ingress.Run(context.Background())

	mockMempool.On("IsKnownOp", mock.Anything).Return(false).Times(3)
	mockMempool.On("AddOperation", mock.Anything, top1, false).Run(func(args mock.Arguments) {
		done <- true
	}).Return(nil).Once()
	mockMempool.On("AddOperation", mock.Anything, top2, false).Run(func(args mock.Arguments) {
		done <- true
	}).Return(nil).Once()
	mockMempool.On("AddOperation", mock.Anything, top3, false).Run(func(args mock.Arguments) {
		done <- true
	}).Return(nil).Once()
	mockCollector.On("ValidatePayment", mock.Anything).Return(nil).Times(3)

	for len(mockP2p.Handlers) == 0 {
	}

	go mockP2p.ExtBroadcast("", proto.MessageType_NEW_OPERATION, op1)
	go mockP2p.ExtBroadcast("", proto.MessageType_NEW_OPERATION, op2)
	go mockP2p.ExtBroadcast("", proto.MessageType_NEW_OPERATION, op3)

	<-done
	<-done
	<-done

	assert.Equal(t, 0, ingress.InBuffer())
}
