package mocks

import (
	"math/big"

	"github.com/0xsequence/bundler/collector"
	"github.com/0xsequence/bundler/pricefeed"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/types"
	"github.com/stretchr/testify/mock"
)

type MockCollector struct {
	mock.Mock
}

func (m *MockCollector) Cmp(a *types.Operation, b *types.Operation) int {
	args := m.Called(a, b)
	return args.Int(0)
}

func (m *MockCollector) Feed(token string) (pricefeed.Feed, error) {
	args := m.Called(token)
	if args.Get(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(pricefeed.Feed), nil
}

func (m *MockCollector) NativeFeesPerGas(a *types.Operation) (*big.Int, *big.Int) {
	args := m.Called(a)
	return args.Get(0).(*big.Int), args.Get(1).(*big.Int)
}

func (m *MockCollector) FeeAsks() (*proto.FeeAsks, error) {
	args := m.Called()
	return args.Get(0).(*proto.FeeAsks), args.Error(1)
}

func (m *MockCollector) BaseFee() *big.Int {
	return m.Called().Get(0).(*big.Int)
}

func (m *MockCollector) Feeds() []pricefeed.Feed {
	return m.Called().Get(0).([]pricefeed.Feed)
}

func (m *MockCollector) ValidatePayment(op *types.Operation) error {
	args := m.Called(op)
	return args.Error(0)
}

func (m *MockCollector) PriorityFee() *big.Int {
	return m.Called().Get(0).(*big.Int)
}

var _ collector.Interface = &MockCollector{}
