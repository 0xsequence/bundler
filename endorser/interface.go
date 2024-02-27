package endorser

import (
	"context"
	"math/big"

	"github.com/0xsequence/bundler/contracts/gen/solabis/abiendorser"
	"github.com/0xsequence/bundler/types"
)

type DependencyState struct {
	Balance *big.Int   `json:"balance,omitempty"`
	Code    []byte     `json:"code,omitempty"`
	Nonce   *uint64    `json:"nonce,omitempty"`
	Slots   [][32]byte `json:"slots,omitempty"`
}

type GlobalDependency = abiendorser.EndorserGlobalDependency
type Dependency = abiendorser.EndorserDependency
type Constraint = abiendorser.EndorserConstraint

type EndorserResult struct {
	Readiness        bool             `json:"readiness"`
	GlobalDependency GlobalDependency `json:"global_dependency"`
	Dependencies     []Dependency     `json:"dependencies"`
}

type EndorserResultState struct {
	Dependencies []DependencyState `json:"dependencies"`
}

type Interface interface {
	IsOperationReady(ctx context.Context, op *types.Operation) (*EndorserResult, error)
	DependencyState(ctx context.Context, result *EndorserResult) (*EndorserResultState, error)
	ConstraintsMet(ctx context.Context, result *EndorserResult) (bool, error)
}
