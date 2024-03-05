package mocks

import (
	"math/big"

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
	_entrypoint common.Address,
	_data []byte,
	_endorserCallData []byte,
	_gasLimit *big.Int,
	_maxFeePerGas *big.Int,
	_maxPriorityFeePerGas *big.Int,
	_feeToken common.Address,
	_baseFeeScalingFactor *big.Int,
	_baseFeeNormalizationFactor *big.Int,
	_hasUntrustedContext bool,
	_endorser common.Address,
) (abivalidator.OperationValidatorSimulationResult, error) {
	args := m.Called(
		opts,
		_entrypoint,
		_data,
		_endorserCallData,
		_gasLimit,
		_maxFeePerGas,
		_maxPriorityFeePerGas,
		_feeToken,
		_baseFeeScalingFactor,
		_baseFeeNormalizationFactor,
		_hasUntrustedContext,
		_endorser,
	)
	err := args.Error(1)
	if err != nil {
		return abivalidator.OperationValidatorSimulationResult{}, err
	}
	return args.Get(0).(abivalidator.OperationValidatorSimulationResult), nil
}

var _ sender.ValidatorInterface = &MockValidator{}
