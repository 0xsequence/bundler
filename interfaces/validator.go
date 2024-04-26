package interfaces

import (
	"github.com/0xsequence/bundler/contracts/gen/solabis/abivalidator"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi/bind"
)

type Validator interface {
	SimulateOperation(opts *bind.CallOpts, _op abivalidator.IEndorserOperation) (abivalidator.OperationValidatorSimulationResult, error)
}

var _ Validator = &abivalidator.OperationValidator{}
