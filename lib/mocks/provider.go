package mocks

import (
	"context"
	"math/big"

	"github.com/0xsequence/ethkit/ethrpc"
	ethereum "github.com/0xsequence/ethkit/go-ethereum"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
)

type MockRPCProvider struct {
	mock.Mock
}

func (m *MockRPCProvider) Do(ctx context.Context, calls ...ethrpc.Call) ([]byte, error) {
	args := m.Called(ctx, calls)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockRPCProvider) NonceAt(ctx context.Context, account common.Address, blockNum *big.Int) (uint64, error) {
	args := m.Called(ctx, account, blockNum)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockRPCProvider) PeerCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockRPCProvider) PendingBalanceAt(ctx context.Context, account common.Address) (*big.Int, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockRPCProvider) TransactionSender(ctx context.Context, tx *types.Transaction, block common.Hash, index uint) (common.Address, error) {
	args := m.Called(ctx, tx, block, index)
	return args.Get(0).(common.Address), args.Error(1)
}

func (m *MockRPCProvider) BalanceAt(ctx context.Context, account common.Address, blockNum *big.Int) (*big.Int, error) {
	args := m.Called(ctx, account, blockNum)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockRPCProvider) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	args := m.Called(ctx, hash)
	return args.Get(0).(*types.Block), args.Error(1)
}

func (m *MockRPCProvider) BlockByNumber(ctx context.Context, blockNum *big.Int) (*types.Block, error) {
	args := m.Called(ctx, blockNum)
	return args.Get(0).(*types.Block), args.Error(1)
}

func (m *MockRPCProvider) BlockNumber(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockRPCProvider) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNum *big.Int) ([]byte, error) {
	args := m.Called(ctx, msg, blockNum)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockRPCProvider) CallContractAtHash(ctx context.Context, msg ethereum.CallMsg, blockHash common.Hash) ([]byte, error) {
	args := m.Called(ctx, msg, blockHash)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockRPCProvider) ChainID(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockRPCProvider) CodeAt(ctx context.Context, account common.Address, blockNum *big.Int) ([]byte, error) {
	args := m.Called(ctx, account, blockNum)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockRPCProvider) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	args := m.Called(ctx, msg)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockRPCProvider) FeeHistory(ctx context.Context, blockCount uint64, lastBlock *big.Int, rewardPercentiles []float64) (*ethereum.FeeHistory, error) {
	args := m.Called(ctx, blockCount, lastBlock, rewardPercentiles)
	return args.Get(0).(*ethereum.FeeHistory), args.Error(1)
}

func (m *MockRPCProvider) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	args := m.Called(ctx, q)
	return args.Get(0).([]types.Log), args.Error(1)
}

func (m *MockRPCProvider) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	args := m.Called(ctx, hash)
	return args.Get(0).(*types.Header), args.Error(1)
}

func (m *MockRPCProvider) HeaderByNumber(ctx context.Context, blockNum *big.Int) (*types.Header, error) {
	args := m.Called(ctx, blockNum)
	return args.Get(0).(*types.Header), args.Error(1)
}

func (m *MockRPCProvider) NetworkID(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockRPCProvider) PendingCallContract(ctx context.Context, msg ethereum.CallMsg) ([]byte, error) {
	args := m.Called(ctx, msg)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockRPCProvider) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	args := m.Called(ctx, account)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockRPCProvider) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockRPCProvider) PendingStorageAt(ctx context.Context, account common.Address, key common.Hash) ([]byte, error) {
	args := m.Called(ctx, account, key)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockRPCProvider) PendingTransactionCount(ctx context.Context) (uint, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint), args.Error(1)
}

func (m *MockRPCProvider) SendRawTransaction(ctx context.Context, signedTxHex string) (common.Hash, error) {
	args := m.Called(ctx, signedTxHex)
	return args.Get(0).(common.Hash), args.Error(1)
}

func (m *MockRPCProvider) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *MockRPCProvider) StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNum *big.Int) ([]byte, error) {
	args := m.Called(ctx, account, key, blockNum)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockRPCProvider) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockRPCProvider) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockRPCProvider) SyncProgress(ctx context.Context) (*ethereum.SyncProgress, error) {
	args := m.Called(ctx)
	return args.Get(0).(*ethereum.SyncProgress), args.Error(1)
}

func (m *MockRPCProvider) TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, pending bool, err error) {
	args := m.Called(ctx, hash)
	return args.Get(0).(*types.Transaction), args.Bool(1), args.Error(2)
}

func (m *MockRPCProvider) TransactionCount(ctx context.Context, blockHash common.Hash) (uint, error) {
	args := m.Called(ctx, blockHash)
	return args.Get(0).(uint), args.Error(1)
}

func (m *MockRPCProvider) TransactionInBlock(ctx context.Context, blockHash common.Hash, index uint) (*types.Transaction, error) {
	args := m.Called(ctx, blockHash, index)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockRPCProvider) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	args := m.Called(ctx, txHash)
	return args.Get(0).(*types.Receipt), args.Error(1)
}

func (m *MockRPCProvider) IsStreamingEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockRPCProvider) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	args := m.Called(ctx, query, ch)
	return args.Get(0).(ethereum.Subscription), args.Error(1)
}

func (m *MockRPCProvider) SubscribeNewHeads(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	args := m.Called(ctx, ch)
	return args.Get(0).(ethereum.Subscription), args.Error(1)
}

var _ ethrpc.Interface = &MockRPCProvider{}
