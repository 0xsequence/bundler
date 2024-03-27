package mocks

import (
	"context"
	"math/big"

	"github.com/0xsequence/bundler/interfaces"
	ethereum "github.com/0xsequence/ethkit/go-ethereum"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
)

type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) BalanceAt(ctx context.Context, account common.Address, blockNum *big.Int) (*big.Int, error) {
	args := m.Called(ctx, account, blockNum)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockProvider) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNum *big.Int) ([]byte, error) {
	args := m.Called(ctx, msg, blockNum)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockProvider) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, contract, blockNumber)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockProvider) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	args := m.Called(ctx, msg)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockProvider) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	args := m.Called(ctx, hash)
	return args.Get(0).(*types.Block), args.Error(1)
}

var _ interfaces.Provider = &MockProvider{}
