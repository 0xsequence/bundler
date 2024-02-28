package mocks

import (
	"context"
	"math/big"

	"github.com/0xsequence/bundler/pricefeed"
	"github.com/stretchr/testify/mock"
)

type MockFeed struct {
	mock.Mock
}

func (m *MockFeed) Factors() (*big.Int, *big.Int, error) {
	args := m.Called()
	err := args.Error(2)
	if err != nil {
		return nil, nil, err
	}

	return args.Get(0).(*big.Int), args.Get(1).(*big.Int), nil
}

func (m *MockFeed) Ready() bool {
	return m.Called().Bool(0)
}

func (m *MockFeed) Name() string {
	return m.Called().String(0)
}

func (m *MockFeed) FromNative(amount *big.Int) (*big.Int, error) {
	args := m.Called(amount)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockFeed) ToNative(amount *big.Int) (*big.Int, error) {
	args := m.Called(amount)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockFeed) Start(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

var _ pricefeed.Feed = &MockFeed{}
