package sender_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/contracts/gen/solabis/abiendorser"
	"github.com/0xsequence/bundler/contracts/gen/solabis/abivalidator"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/lib/collector"
	"github.com/0xsequence/bundler/lib/mempool"
	"github.com/0xsequence/bundler/lib/mocks"
	"github.com/0xsequence/bundler/lib/pricefeed"
	"github.com/0xsequence/bundler/lib/registry"
	"github.com/0xsequence/bundler/lib/types"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/sender"
	"github.com/0xsequence/ethkit/ethtxn"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/go-chi/httplog/v2"
	"github.com/stretchr/testify/mock"

	ethtypes "github.com/0xsequence/ethkit/go-ethereum/core/types"
)

func TestReservePullOps(t *testing.T) {
	logger := httplog.NewLogger("")
	mockWallet := &mocks.MockWallet{}
	mockWalletFactory := &mocks.MockWalletFactory{}
	mockValidator := &mocks.MockValidator{}
	mockMempool := &mocks.MockMempool{}
	mockEndorser := &mocks.MockEndorser{}
	mockProvider := &mocks.MockProvider{}
	mockCollector := &mocks.MockCollector{}
	mockRegistry := &mocks.MockRegistry{}

	mockWallet.On("Address").Return(common.Address{}, nil).Maybe()
	mockWalletFactory.On("GetWallet", mock.Anything).Return(mockWallet, nil).Maybe()
	mockProvider.On("BalanceAt", mock.Anything, mock.Anything, mock.Anything).Return(big.NewInt(1000000000000000000), nil).Maybe()

	sender := sender.NewSender(
		&config.SendersConfig{
			SleepWait:  1,
			NumSenders: 1,
		},
		logger,
		mockWalletFactory,
		mockProvider,
		mockMempool,
		mockEndorser,
		mockValidator,
		mockCollector,
		mockRegistry,
	)

	done := make(chan struct{})
	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			<-done
		}).
		Return([]*mempool.TrackedOperation{}, nil).Once()
	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).Return([]*mempool.TrackedOperation{}, nil).Maybe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sender.Run(ctx)

	done <- struct{}{}
}

func TestSimulateOpErr(t *testing.T) {
	logger := httplog.NewLogger("")
	mockWallet := &mocks.MockWallet{}
	mockWalletFactory := &mocks.MockWalletFactory{}
	mockValidator := &mocks.MockValidator{}
	mockMempool := &mocks.MockMempool{}
	mockEndorser := &mocks.MockEndorser{}
	mockProvider := &mocks.MockProvider{}
	mockCollector := &mocks.MockCollector{}
	mockRegistry := &mocks.MockRegistry{}

	addr := common.HexToAddress("0x7537713a54d2506b36eFa389F9341d63815ddE48")
	balance := big.NewInt(1000000000000000000)

	mockWallet.On("Address").Return(addr, nil).Maybe()
	mockWalletFactory.On("GetWallet", mock.Anything).Return(mockWallet, nil).Maybe()
	mockProvider.On("BalanceAt", mock.Anything, mock.Anything, mock.Anything).Return(balance, nil).Maybe()
	mockProvider.On("EstimateGas", mock.Anything, mock.Anything).Return(uint64(100000), nil).Maybe()

	op := mempool.TrackedOperation{
		Operation: types.Operation{},
	}

	sender := sender.NewSender(
		&config.SendersConfig{
			SleepWait:  1,
			NumSenders: 1,
		},
		logger,
		mockWalletFactory,
		mockProvider,
		mockMempool,
		mockEndorser,
		mockValidator,
		mockCollector,
		mockRegistry,
	)

	done := make(chan struct{})

	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).Return([]*mempool.TrackedOperation{&op}, nil).Once()

	mockValidator.On(
		"SimulateOperation",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil, fmt.Errorf("err")).Once()

	mockMempool.On("ReleaseOps", mock.Anything, mock.Anything, mock.Anything).
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

	cancel()
}

func TestSimulatePaidNotPaid(t *testing.T) {
	logger := httplog.NewLogger("")
	mockWallet := &mocks.MockWallet{}
	mockWalletFactory := &mocks.MockWalletFactory{}
	mockValidator := &mocks.MockValidator{}
	mockMempool := &mocks.MockMempool{}
	mockEndorser := &mocks.MockEndorser{}
	mockProvider := &mocks.MockProvider{}
	mockCollector := &mocks.MockCollector{}
	mockRegistry := &mocks.MockRegistry{}

	mockWallet.On("Address").Return(common.Address{}, nil).Maybe()
	mockWalletFactory.On("GetWallet", mock.Anything).Return(mockWallet, nil).Maybe()
	mockProvider.On("EstimateGas", mock.Anything, mock.Anything).Return(uint64(100000), nil).Maybe()
	mockProvider.On("BalanceAt", mock.Anything, mock.Anything, mock.Anything).Return(big.NewInt(1000000000000000000), nil).Maybe()
	mockCollector.On("BaseFee").Return(big.NewInt(2), nil).Maybe()
	mockCollector.On("NativeFeesPerGas", mock.Anything).Return(&collector.NativeFees{}, &pricefeed.Snapshot{
		ScalingFactor:       big.NewInt(1),
		NormalizationFactor: big.NewInt(1),
	})

	op := mempool.TrackedOperation{
		Operation: types.Operation{},
	}

	sender := sender.NewSender(
		&config.SendersConfig{
			SleepWait:  1,
			NumSenders: 1,
		},
		logger,
		mockWalletFactory,
		mockProvider,
		mockMempool,
		mockEndorser,
		mockValidator,
		mockCollector,
		mockRegistry,
	)

	done := make(chan struct{})

	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).Return([]*mempool.TrackedOperation{&op}, nil).Once()

	mockValidator.On(
		"SimulateOperation",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(abivalidator.OperationValidatorSimulationResult{
		Payment: big.NewInt(100),
		GasUsed: big.NewInt(300000),
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

func TestSendAndBanEndorserFailedTx(t *testing.T) {
	logger := httplog.NewLogger("")
	mockWallet := &mocks.MockWallet{}
	mockWalletFactory := &mocks.MockWalletFactory{}
	mockValidator := &mocks.MockValidator{}
	mockMempool := &mocks.MockMempool{}
	mockEndorser := &mocks.MockEndorser{}
	mockProvider := &mocks.MockProvider{}
	mockCollector := &mocks.MockCollector{}
	mockRegistry := &mocks.MockRegistry{}

	endorserAddr := common.HexToAddress("0x08FFc248A190E700421C0aFB4135768406dCebfF")

	mockWallet.On("Address").Return(common.Address{}, nil).Maybe()
	mockWalletFactory.On("GetWallet", mock.Anything).Return(mockWallet, nil).Twice()
	mockProvider.On("EstimateGas", mock.Anything, mock.Anything).Return(uint64(100000), nil).Twice()
	mockProvider.On("BalanceAt", mock.Anything, mock.Anything, mock.Anything).Return(big.NewInt(1000000000000000000), nil).Maybe()
	mockCollector.On("BaseFee").Return(big.NewInt(213), nil).Maybe()
	mockCollector.On("NativeFeesPerGas", mock.Anything).Return(&collector.NativeFees{}, &pricefeed.Snapshot{
		ScalingFactor:       big.NewInt(1),
		NormalizationFactor: big.NewInt(1),
	}).Maybe()

	op := mempool.TrackedOperation{
		Operation: types.Operation{
			Endorser: endorserAddr,
			IEndorserOperation: abiendorser.IEndorserOperation{
				GasLimit:             big.NewInt(1000),
				MaxFeePerGas:         big.NewInt(213),
				MaxPriorityFeePerGas: big.NewInt(50),
				Entrypoint:           common.HexToAddress("0xB0e4BDF60bC80cbCAaC52DF8796e579870d2fd00"),
				Data:                 common.Hex2Bytes("0x1234"),
			},
		},
	}

	sender := sender.NewSender(
		&config.SendersConfig{
			SleepWait:   1,
			NumSenders:  1,
			PriorityFee: 13,
		},
		logger,
		mockWalletFactory,
		mockProvider,
		mockMempool,
		mockEndorser,
		mockValidator,
		mockCollector,
		mockRegistry,
	)

	done := make(chan struct{})

	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).Return([]*mempool.TrackedOperation{&op}, nil).Once()
	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).Return([]*mempool.TrackedOperation{}, nil).Maybe()

	mockValidator.On(
		"SimulateOperation",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(abivalidator.OperationValidatorSimulationResult{
		Payment: big.NewInt(1000000000000000000),
		GasUsed: big.NewInt(100000),
	}, nil).Once()

	mockEndorser.On("IsOperationReady", mock.Anything, mock.Anything).Return(&endorser.EndorserResult{
		Readiness:    true,
		Dependencies: []abiendorser.IEndorserDependency{{}},
	}, nil).Once()

	mockEndorser.On("ConstraintsMet", mock.Anything, mock.Anything).Return(true, nil).Once()

	rtx := ethtypes.Transaction{}
	mockWallet.On("NewTransaction", mock.Anything, &ethtxn.TransactionRequest{
		To:       &op.Operation.Entrypoint,
		GasPrice: big.NewInt(226),
		GasTip:   big.NewInt(13),
		GasLimit: 101000,
		Data:     op.Operation.Data,
		ETHValue: big.NewInt(0),
	}).Return(&rtx, nil).Once()

	var waitFn ethtxn.WaitReceipt
	_ = waitFn

	waitFn = func(context.Context) (*ethtypes.Receipt, error) {
		return &ethtypes.Receipt{
			Status:            0,
			TxHash:            common.HexToHash("0x1234"),
			BlockNumber:       big.NewInt(100),
			EffectiveGasPrice: big.NewInt(213),
		}, nil
	}

	mockWallet.On("SendTransaction", mock.Anything, &rtx).Return(&rtx, waitFn, nil).Once()
	mockMempool.On("ReleaseOps", mock.Anything, mock.Anything, proto.ReadyAtChange_None).Return(nil).Maybe()
	mockProvider.On("EstimateGas", mock.Anything, mock.Anything).Return(uint64(10), nil).Once()
	mockCollector.On("NativeFeesPerGas", &op.Operation).Return(&collector.NativeFees{
		MaxFeePerGas:         big.NewInt(213),
		MaxPriorityFeePerGas: big.NewInt(50),
	}, &pricefeed.Snapshot{}).Once()
	mockCollector.On("BaseFee").Return(big.NewInt(100), nil).Once()

	mockRegistry.On("BanEndorser", endorserAddr, registry.PermanentBan).Run(func(args mock.Arguments) {
		done <- struct{}{}
	}).Return().Once()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sender.Run(ctx)

	<-done

	mockWallet.AssertExpectations(t)
	mockMempool.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

func TestSendAndBanEndorserLowPayment(t *testing.T) {
	logger := httplog.NewLogger("")
	mockWallet := &mocks.MockWallet{}
	mockWalletFactory := &mocks.MockWalletFactory{}
	mockValidator := &mocks.MockValidator{}
	mockMempool := &mocks.MockMempool{}
	mockEndorser := &mocks.MockEndorser{}
	mockProvider := &mocks.MockProvider{}
	mockCollector := &mocks.MockCollector{}
	mockRegistry := &mocks.MockRegistry{}

	endorserAddr := common.HexToAddress("0x08FFc248A190E700421C0aFB4135768406dCebfF")

	mockWallet.On("Address").Return(common.Address{}, nil).Maybe()
	mockWalletFactory.On("GetWallet", mock.Anything).Return(mockWallet, nil).Twice()
	mockProvider.On("EstimateGas", mock.Anything, mock.Anything).Return(uint64(100000), nil).Twice()
	mockProvider.On("BalanceAt", mock.Anything, mock.Anything, mock.Anything).Return(big.NewInt(1000000000000000000), nil).Maybe()
	mockCollector.On("BaseFee").Return(big.NewInt(213), nil).Maybe()
	mockCollector.On("NativeFeesPerGas", mock.Anything).Return(&collector.NativeFees{}, &pricefeed.Snapshot{
		ScalingFactor:       big.NewInt(1),
		NormalizationFactor: big.NewInt(1),
	}).Maybe()

	op := mempool.TrackedOperation{
		Operation: types.Operation{
			Endorser: endorserAddr,
			IEndorserOperation: abiendorser.IEndorserOperation{
				GasLimit:             big.NewInt(1000),
				MaxFeePerGas:         big.NewInt(213),
				MaxPriorityFeePerGas: big.NewInt(50),
				Entrypoint:           common.HexToAddress("0xB0e4BDF60bC80cbCAaC52DF8796e579870d2fd00"),
				Data:                 common.Hex2Bytes("0x1234"),
			},
		},
	}

	sender := sender.NewSender(
		&config.SendersConfig{
			SleepWait:   1,
			PriorityFee: 13,
			NumSenders:  1,
		},
		logger,
		mockWalletFactory,
		mockProvider,
		mockMempool,
		mockEndorser,
		mockValidator,
		mockCollector,
		mockRegistry,
	)

	done := make(chan struct{})

	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).Return([]*mempool.TrackedOperation{&op}, nil).Once()
	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).Return([]*mempool.TrackedOperation{}, nil).Maybe()

	mockValidator.On(
		"SimulateOperation",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(abivalidator.OperationValidatorSimulationResult{
		Payment: big.NewInt(1000000000000000000),
		GasUsed: big.NewInt(100000),
	}, nil).Once()

	rtx := ethtypes.Transaction{}
	mockWallet.On("NewTransaction", mock.Anything, &ethtxn.TransactionRequest{
		To:       &op.Operation.Entrypoint,
		GasPrice: big.NewInt(226),
		GasTip:   big.NewInt(13),
		GasLimit: 101000,
		Data:     op.Operation.Data,
		ETHValue: big.NewInt(0),
	}).Return(&rtx, nil).Once()

	var waitFn ethtxn.WaitReceipt
	_ = waitFn

	waitFn = func(context.Context) (*ethtypes.Receipt, error) {
		return &ethtypes.Receipt{
			Status:            1,
			TxHash:            common.HexToHash("0x1234"),
			BlockNumber:       big.NewInt(100),
			EffectiveGasPrice: big.NewInt(213),
			GasUsed:           10,
		}, nil
	}

	mockWallet.On("SendTransaction", mock.Anything, &rtx).Return(&rtx, waitFn, nil).Once()
	mockMempool.On("ReleaseOps", mock.Anything, mock.Anything, proto.ReadyAtChange_None).Return(nil).Maybe()
	mockProvider.On("EstimateGas", mock.Anything, mock.Anything).Return(uint64(10), nil).Once()
	mockCollector.On("NativeFeesPerGas", &op.Operation).Return(&collector.NativeFees{
		MaxFeePerGas:         big.NewInt(213),
		MaxPriorityFeePerGas: big.NewInt(50),
	}, &pricefeed.Snapshot{}).Once()
	mockCollector.On("BaseFee").Return(big.NewInt(100), nil).Once()

	mockRegistry.On("BanEndorser", endorserAddr, registry.PermanentBan).Run(func(args mock.Arguments) {
		done <- struct{}{}
	}).Return().Once()

	mockProvider.On("BalanceAt", mock.Anything, mock.Anything, big.NewInt(99)).Return(big.NewInt(10), nil).Once()
	mockProvider.On("BalanceAt", mock.Anything, mock.Anything, big.NewInt(100)).Return(big.NewInt(9), nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sender.Run(ctx)

	<-done

	mockWallet.AssertExpectations(t)
	mockMempool.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

func TestSend(t *testing.T) {
	logger := httplog.NewLogger("")
	mockWallet := &mocks.MockWallet{}
	mockWalletFactory := &mocks.MockWalletFactory{}
	mockValidator := &mocks.MockValidator{}
	mockMempool := &mocks.MockMempool{}
	mockEndorser := &mocks.MockEndorser{}
	mockProvider := &mocks.MockProvider{}
	mockCollector := &mocks.MockCollector{}
	mockRegistry := &mocks.MockRegistry{}

	addr := common.HexToAddress("0x7537713a54d2506b36eFa389F9341d63815ddE48")
	mockWallet.On("Address").Return(addr, nil).Maybe()
	mockWalletFactory.On("GetWallet", mock.Anything).Return(mockWallet, nil).Twice()
	mockProvider.On("EstimateGas", mock.Anything, mock.Anything).Return(uint64(100000), nil).Twice()
	mockProvider.On("BalanceAt", mock.Anything, addr, mock.Anything).Return(big.NewInt(2000000000000000000), nil).Once()
	mockCollector.On("BaseFee").Return(big.NewInt(213), nil).Maybe()
	mockCollector.On("NativeFeesPerGas", mock.Anything).Return(&collector.NativeFees{}, &pricefeed.Snapshot{
		ScalingFactor:       big.NewInt(1),
		NormalizationFactor: big.NewInt(1),
	}).Maybe()

	op := mempool.TrackedOperation{
		Operation: types.Operation{
			IEndorserOperation: abiendorser.IEndorserOperation{
				GasLimit:             big.NewInt(1000),
				MaxFeePerGas:         big.NewInt(213),
				MaxPriorityFeePerGas: big.NewInt(50),
				Entrypoint:           common.HexToAddress("0xB0e4BDF60bC80cbCAaC52DF8796e579870d2fd00"),
				Data:                 common.Hex2Bytes("0x1234"),
			},
		},
	}

	sender := sender.NewSender(
		&config.SendersConfig{
			SleepWait:   1,
			PriorityFee: 13,
			NumSenders:  1,
		},
		logger,
		mockWalletFactory,
		mockProvider,
		mockMempool,
		mockEndorser,
		mockValidator,
		mockCollector,
		mockRegistry,
	)

	done := make(chan struct{})

	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).Return([]*mempool.TrackedOperation{&op}, nil).Once()
	mockMempool.On("ReserveOps", mock.Anything, mock.Anything).Return([]*mempool.TrackedOperation{}, nil).Maybe()

	mockValidator.On(
		"SimulateOperation",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(abivalidator.OperationValidatorSimulationResult{
		Payment: big.NewInt(1000000000000000000),
		GasUsed: big.NewInt(100000),
	}, nil).Once()

	rtx := ethtypes.Transaction{}
	mockWallet.On("NewTransaction", mock.Anything, &ethtxn.TransactionRequest{
		To:       &op.Operation.Entrypoint,
		GasPrice: big.NewInt(226),
		GasTip:   big.NewInt(13),
		GasLimit: 101000,
		Data:     op.Operation.Data,
		ETHValue: big.NewInt(0),
	}).Return(&rtx, nil).Once()

	var waitFn ethtxn.WaitReceipt
	_ = waitFn

	waitFn = func(context.Context) (*ethtypes.Receipt, error) {
		return &ethtypes.Receipt{
			Status:            1,
			TxHash:            common.HexToHash("0x1234"),
			BlockNumber:       big.NewInt(100),
			EffectiveGasPrice: big.NewInt(213),
		}, nil
	}

	mockWallet.On("SendTransaction", mock.Anything, &rtx).Return(&rtx, waitFn, nil).Once()
	mockMempool.On("ReleaseOps", mock.Anything, mock.Anything, proto.ReadyAtChange_None).
		Run(func(args mock.Arguments) {
			done <- struct{}{}
		}).Return(nil).Once()

	mockProvider.On("EstimateGas", mock.Anything, mock.Anything).Return(uint64(10), nil).Once()
	mockCollector.On("NativeFeesPerGas", &op.Operation).Return(&collector.NativeFees{
		MaxFeePerGas:         big.NewInt(213),
		MaxPriorityFeePerGas: big.NewInt(50),
	}, &pricefeed.Snapshot{}).Once()
	mockCollector.On("BaseFee").Return(big.NewInt(100), nil).Once()
	mockProvider.On("BalanceAt", mock.Anything, mock.Anything, big.NewInt(99)).Return(big.NewInt(2000000000000000000), nil).Once()
	mockProvider.On("BalanceAt", mock.Anything, mock.Anything, big.NewInt(100)).Return(big.NewInt(4000000000000000000), nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sender.Run(ctx)

	<-done

	mockWallet.AssertExpectations(t)
	mockMempool.AssertExpectations(t)
	mockValidator.AssertExpectations(t)

	// Delay 100 ms to inspect receipt
	time.Sleep(100 * time.Millisecond)
}
