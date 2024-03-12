package mocks

import (
	"context"

	"github.com/0xsequence/bundler/sender"
	ethereum "github.com/0xsequence/ethkit/go-ethereum"
	"github.com/stretchr/testify/mock"
)

type MockGasEstimator struct {
	mock.Mock
}

func (g *MockGasEstimator) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	args := g.Called(ctx, msg)
	return args.Get(0).(uint64), args.Error(1)
}

var _ sender.GasEstimator = &MockGasEstimator{}
