package interfaces

import (
	"github.com/0xsequence/bundler/contracts/gen/solabis/abivalidator2"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi/bind"
)

type Validator2 interface {
	SimulateOperation(opts *bind.CallOpts, _op abivalidator2.IEndorserOperation) (abivalidator2.OperationValidator2SimulationResult2, error)
}

var _ Validator2 = &abivalidator2.OperationValidator2{}
