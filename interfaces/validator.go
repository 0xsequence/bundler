package interfaces

import (
	"github.com/0xsequence/bundler/contracts/gen/solabis/abivalidator"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi/bind"
	"github.com/0xsequence/ethkit/go-ethereum/common"
)

type Validator interface {
	SimulateOperation(opts *bind.CallOpts, _endorser common.Address, _op abivalidator.IEndorserOperation) (abivalidator.OperationValidatorSimulationResult, error)
}

var _ Validator = &abivalidator.OperationValidator{}
