package mocks

import (
	"math/big"

	"github.com/0xsequence/bundler/collector"
	"github.com/0xsequence/bundler/pricefeed"
	"github.com/0xsequence/bundler/types"
	"github.com/stretchr/testify/mock"
)

type MockCollector struct {
	mock.Mock
}

func (m *MockCollector) BaseFee() *big.Int {
	return m.Called().Get(0).(*big.Int)
}

func (m *MockCollector) Feeds() []pricefeed.Feed {
	return m.Called().Get(0).([]pricefeed.Feed)
}

func (m *MockCollector) MeetsPayment(op *types.Operation) (bool, error) {
	args := m.Called(op)
	return args.Bool(0), args.Error(1)
}

func (m *MockCollector) PriorityFee() *big.Int {
	return m.Called().Get(0).(*big.Int)
}

var _ collector.Interface = &MockCollector{}
