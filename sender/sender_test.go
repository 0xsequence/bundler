package sender_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/0xsequence/bundler/calldata"
	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/contracts/gen/solabis/abiendorser"
	"github.com/0xsequence/bundler/contracts/gen/solabis/abivalidator"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/mocks"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/sender"
	"github.com/0xsequence/bundler/types"
	"github.com/0xsequence/ethkit/ethtxn"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	ethtypes "github.com/0xsequence/ethkit/go-ethereum/core/types"
	"github.com/go-chi/httplog/v2"
	"github.com/stretchr/testify/mock"
)

func TestReservePullOps(t *testing.T) {
	logger := httplog.NewLogger("").With("", "")
	mockWallet := &mocks.MockWallet{}
	mockValidator := &mocks.MockValidator{}
	mockMempool := &mocks.MockMempool{}
	mockEndorser := &mocks.MockEndorser{}

	sender := sender.NewSender(
		&config.SendersConfig{
			SleepWait: 1,
		},
		logger,
		0,
		mockWallet,
		mockMempool,
		mockEndorser,
		mockValidator,
		calldata.DefaultModel(),
	)

	done := make(chan struct{})
	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			<-done
		}).
		Return([]*mempool.TrackedOperation{}, nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sender.Run(ctx)

	done <- struct{}{}
}

func TestSimulateOpErr(t *testing.T) {
	logger := httplog.NewLogger("").With("", "")
	mockWallet := &mocks.MockWallet{}
	mockValidator := &mocks.MockValidator{}
	mockMempool := &mocks.MockMempool{}
	mockEndorser := &mocks.MockEndorser{}

	op := mempool.TrackedOperation{
		Operation: types.Operation{},
	}

	sender := sender.NewSender(
		&config.SendersConfig{
			SleepWait: 1,
		},
		logger,
		0,
		mockWallet,
		mockMempool,
		mockEndorser,
		mockValidator,
		calldata.DefaultModel(),
	)

	done := make(chan struct{})

	mockWallet.On("Address").Return(common.Address{}, nil).Once()

	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).Return([]*mempool.TrackedOperation{&op}, nil).Once()

	mockValidator.On(
		"SimulateOperation",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil, fmt.Errorf("err")).Once()

	mockMempool.On("DiscardOps", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			done <- struct{}{}
		}).
		Return(nil).Once()

	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).Return([]*mempool.TrackedOperation{}, nil).Maybe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sender.Run(ctx)

	<-done

	mockWallet.AssertExpectations(t)
	mockMempool.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

func TestSimulatePaidNotPaid(t *testing.T) {
	logger := httplog.NewLogger("").With("", "")
	mockWallet := &mocks.MockWallet{}
	mockValidator := &mocks.MockValidator{}
	mockMempool := &mocks.MockMempool{}
	mockEndorser := &mocks.MockEndorser{}

	op := mempool.TrackedOperation{
		Operation: types.Operation{},
	}

	sender := sender.NewSender(
		&config.SendersConfig{
			SleepWait: 1,
		},
		logger,
		0,
		mockWallet,
		mockMempool,
		mockEndorser,
		mockValidator,
		calldata.DefaultModel(),
	)

	done := make(chan struct{})

	mockWallet.On("Address").Return(common.Address{}, nil).Once()

	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).Return([]*mempool.TrackedOperation{&op}, nil).Once()

	mockValidator.On(
		"SimulateOperation",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(abivalidator.OperationValidatorSimulationResult{
		Paid:      false,
		Readiness: false,
	}, nil).Once()

	mockMempool.On("DiscardOps", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			done <- struct{}{}
		}).
		Return(nil).Once()

	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).Return([]*mempool.TrackedOperation{}, nil).Maybe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sender.Run(ctx)

	<-done

	mockWallet.AssertExpectations(t)
	mockMempool.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

func TestSimulatePaidNotPaidConstraintsUnmet(t *testing.T) {
	logger := httplog.NewLogger("").With("", "")
	mockWallet := &mocks.MockWallet{}
	mockValidator := &mocks.MockValidator{}
	mockMempool := &mocks.MockMempool{}
	mockEndorser := &mocks.MockEndorser{}

	op := mempool.TrackedOperation{
		Operation: types.Operation{},
	}

	sender := sender.NewSender(
		&config.SendersConfig{
			SleepWait: 1,
		},
		logger,
		0,
		mockWallet,
		mockMempool,
		mockEndorser,
		mockValidator,
		calldata.DefaultModel(),
	)

	done := make(chan struct{})

	mockWallet.On("Address").Return(common.Address{}, nil).Once()

	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).Return([]*mempool.TrackedOperation{&op}, nil).Once()

	mockValidator.On(
		"SimulateOperation",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(abivalidator.OperationValidatorSimulationResult{
		Paid:      false,
		Readiness: true,
		Dependencies: []abivalidator.IEndorserDependency{{
			Addr: common.HexToAddress("0x7537713a54d2506b36eFa389F9341d63815ddE48"),
			Constraints: []abivalidator.IEndorserConstraint{{
				Slot:     [32]byte(common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000001")),
				MinValue: [32]byte(common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000002")),
				MaxValue: [32]byte(common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000003")),
			}},
		}},
	}, nil).Once()

	mockEndorser.On(
		"ConstraintsMet",
		mock.Anything,
		&endorser.EndorserResult{
			Readiness:        true,
			GlobalDependency: abiendorser.IEndorserGlobalDependency{},
			Dependencies: []abiendorser.IEndorserDependency{{
				Addr: common.HexToAddress("0x7537713a54d2506b36eFa389F9341d63815ddE48"),
				Constraints: []abiendorser.IEndorserConstraint{{
					Slot:     [32]byte(common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000001")),
					MinValue: [32]byte(common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000002")),
					MaxValue: [32]byte(common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000003")),
				}},
			}},
		},
	).Return(false, nil).Once()

	mockMempool.On("DiscardOps", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			done <- struct{}{}
		}).
		Return(nil).Once()

	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).Return([]*mempool.TrackedOperation{}, nil).Maybe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sender.Run(ctx)

	<-done

	mockWallet.AssertExpectations(t)
	mockMempool.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

func TestSimulatePaidNotPaidAndLied(t *testing.T) {
	logger := httplog.NewLogger("").With("", "")
	mockWallet := &mocks.MockWallet{}
	mockValidator := &mocks.MockValidator{}
	mockMempool := &mocks.MockMempool{}
	mockEndorser := &mocks.MockEndorser{}

	op := mempool.TrackedOperation{
		Operation: types.Operation{},
	}

	sender := sender.NewSender(
		&config.SendersConfig{
			SleepWait: 1,
		},
		logger,
		0,
		mockWallet,
		mockMempool,
		mockEndorser,
		mockValidator,
		calldata.DefaultModel(),
	)

	done := make(chan struct{})

	mockWallet.On("Address").Return(common.Address{}, nil).Once()

	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).Return([]*mempool.TrackedOperation{&op}, nil).Once()

	simErr := "11223344"
	simErr += "0000000000000000000000000000000000000000000000000000000000000001"
	simErr += "000000000000000000000000000000000000000000000000000000000000000a"
	simErr += "000000000000000000000000000000000000000000000000000000000000000b"

	mockValidator.On(
		"SimulateOperation",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(abivalidator.OperationValidatorSimulationResult{
		Paid:         false,
		Readiness:    true,
		Dependencies: []abivalidator.IEndorserDependency{},
		Err:          common.Hex2Bytes(simErr),
	}, nil).Once()

	mockEndorser.On("ConstraintsMet", mock.Anything, mock.Anything).Return(true, nil).Once()

	mockMempool.On("DiscardOps", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			done <- struct{}{}
		}).
		Return(nil).Once()

	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).Return([]*mempool.TrackedOperation{}, nil).Maybe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sender.Run(ctx)

	<-done

	// TODO: Test that endorser is banned

	mockWallet.AssertExpectations(t)
	mockMempool.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

func TestSend(t *testing.T) {
	logger := httplog.NewLogger("").With("", "")
	mockWallet := &mocks.MockWallet{}
	mockValidator := &mocks.MockValidator{}
	mockMempool := &mocks.MockMempool{}
	mockEndorser := &mocks.MockEndorser{}

	op := mempool.TrackedOperation{
		Operation: types.Operation{
			IEndorserOperation: abiendorser.IEndorserOperation{
				GasLimit:     big.NewInt(1000),
				MaxFeePerGas: big.NewInt(2000),
				Entrypoint:   common.HexToAddress("0xB0e4BDF60bC80cbCAaC52DF8796e579870d2fd00"),
				Data:         common.Hex2Bytes("0x1234"),
			},
		},
	}

	sender := sender.NewSender(
		&config.SendersConfig{
			SleepWait:   1,
			PriorityFee: 13,
		},
		logger,
		0,
		mockWallet,
		mockMempool,
		mockEndorser,
		mockValidator,
		calldata.DefaultModel(),
	)

	done := make(chan struct{})

	mockWallet.On("Address").Return(common.Address{}, nil).Once()

	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).Return([]*mempool.TrackedOperation{&op}, nil).Once()
	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).Return([]*mempool.TrackedOperation{}, nil).Maybe()

	mockValidator.On(
		"SimulateOperation",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(abivalidator.OperationValidatorSimulationResult{
		Paid: true,
	}, nil).Once()

	rtx := ethtypes.Transaction{}
	mockWallet.On("NewTransaction", mock.Anything, &ethtxn.TransactionRequest{
		To:       &op.Operation.Entrypoint,
		GasPrice: op.Operation.MaxFeePerGas,
		GasTip:   big.NewInt(13),
		GasLimit: 22000,
		Data:     op.Operation.Data,
		ETHValue: big.NewInt(0),
	}).Return(&rtx, nil).Once()

	var waitFn ethtxn.WaitReceipt
	_ = waitFn

	waitFn = func(context.Context) (*ethtypes.Receipt, error) {
		return &ethtypes.Receipt{
			TxHash: common.HexToHash("0x1234"),
		}, nil
	}

	mockWallet.On("SendTransaction", mock.Anything, &rtx).Return(&rtx, waitFn, nil).Once()
	mockMempool.On("ReleaseOps", mock.Anything, mock.Anything, proto.ReadyAtChange_Zero).
		Run(func(args mock.Arguments) {
			done <- struct{}{}
		}).Return(nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sender.Run(ctx)

	<-done

	mockWallet.AssertExpectations(t)
	mockMempool.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}
