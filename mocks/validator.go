package mocks

import (
	"github.com/0xsequence/bundler/contracts/gen/solabis/abivalidator"
	"github.com/0xsequence/bundler/sender"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi/bind"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/stretchr/testify/mock"
)

type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) SimulateOperation(
	opts *bind.CallOpts,
	_endorser common.Address,
	_operation abivalidator.IEndorserOperation,
) (abivalidator.OperationValidatorSimulationResult, error) {
	args := m.Called(
		opts,
		_endorser,
		_operation,
	)
	err := args.Error(1)
	if err != nil {
		return abivalidator.OperationValidatorSimulationResult{}, err
	}
	return args.Get(0).(abivalidator.OperationValidatorSimulationResult), nil
}

var _ sender.ValidatorInterface = &MockValidator{}
