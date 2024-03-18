package mocks

import (
	"context"

	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/types"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/stretchr/testify/mock"
)

type MockEndorser struct {
	mock.Mock
}

func (m *MockEndorser) ConstraintsMet(ctx context.Context, result *endorser.EndorserResult) (bool, error) {
	args := m.Called(ctx, result)
	return args.Bool(0), args.Error(1)
}

func (m *MockEndorser) DependencyState(ctx context.Context, result *endorser.EndorserResult) (*endorser.EndorserResultState, error) {
	args := m.Called(ctx, result)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*endorser.EndorserResultState), args.Error(1)
}

func (m *MockEndorser) IsOperationReady(ctx context.Context, op *types.Operation) (*endorser.EndorserResult, error) {
	args := m.Called(ctx, op)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*endorser.EndorserResult), args.Error(1)
}

func (m *MockEndorser) SimulationSettings(ctx context.Context, endorserAddr common.Address) ([]*endorser.SimulationSetting, error) {
	args := m.Called(ctx, endorserAddr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*endorser.SimulationSetting), args.Error(1)
}

var _ endorser.Interface = &MockEndorser{}
