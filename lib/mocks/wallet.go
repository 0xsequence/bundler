package mocks

import (
	"context"

	"github.com/0xsequence/bundler/lib/interfaces"
	"github.com/0xsequence/bundler/sender"
	"github.com/0xsequence/ethkit/ethtxn"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
)

type MockWallet struct {
	mock.Mock
}

func (m *MockWallet) Address() common.Address {
	args := m.Called()
	return args.Get(0).(common.Address)
}

func (m *MockWallet) GetNonce(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockWallet) NewTransaction(ctx context.Context, txnRequest *ethtxn.TransactionRequest) (*types.Transaction, error) {
	args := m.Called(ctx, txnRequest)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockWallet) SendTransaction(ctx context.Context, t *types.Transaction) (*types.Transaction, ethtxn.WaitReceipt, error) {
	args := m.Called(ctx, t)
	return args.Get(0).(*types.Transaction), args.Get(1).(ethtxn.WaitReceipt), args.Error(2)
}

var _ interfaces.Wallet = &MockWallet{}

type MockWalletFactory struct {
	mock.Mock
}

func (m *MockWalletFactory) GetWallet(i int) (interfaces.Wallet, error) {
	args := m.Called(i)
	err := args.Error(1)
	if err != nil {
		return nil, err
	}

	return args.Get(0).(interfaces.Wallet), nil
}

var _ sender.WalletFactory = &MockWalletFactory{}
