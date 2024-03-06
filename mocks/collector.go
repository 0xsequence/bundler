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
