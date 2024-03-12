package mocks

import (
	"context"
	"math/big"

	"github.com/0xsequence/bundler/sender"
	ethereum "github.com/0xsequence/ethkit/go-ethereum"
	"github.com/0xsequence/ethkit/go-ethereum/common"
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

var _ sender.Provider = &MockProvider{}
