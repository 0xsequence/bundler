package endorser

import (
	"context"
	"math/big"

	"github.com/0xsequence/bundler/contracts/gen/solabis/abiendorser"
	"github.com/0xsequence/bundler/lib/types"
	"github.com/0xsequence/ethkit/go-ethereum/common"
)

type SlotReplacement struct {
	Slot  [32]byte
	Value [32]byte
}

type SimulationSetting struct {
	OldAddr common.Address
	NewAddr common.Address
	Slots   []SlotReplacement
}

type GlobalDependencyState struct {
	Basefee     *big.Int
	Blobbasefee *big.Int
	Chainid     *big.Int
	Coinbase    common.Address
	Difficulty  *big.Int
	GasLimit    *big.Int
	Number      *big.Int
	Timestamp   *big.Int
	TxOrigin    common.Address
	TxGasPrice  *big.Int
}

type AddrDependencyState struct {
	Balance *big.Int
	Code    *int
	Nonce   *uint64
	Slots   [][32]byte
}

type GlobalDependency = abiendorser.IEndorserGlobalDependency
type Dependency = abiendorser.IEndorserDependency
type Constraint = abiendorser.IEndorserConstraint

type EndorserResult struct {
	WildcardOnly bool

	Readiness        bool             `json:"readiness"`
	GlobalDependency GlobalDependency `json:"global_dependency"`
	Dependencies     []Dependency     `json:"dependencies"`
}

type EndorserResultState struct {
	GlobalDependency *GlobalDependencyState
	AddrDependencies map[common.Address]*AddrDependencyState
}

type Interface interface {
	SimulationSettings(ctx context.Context, endorserAddr common.Address) ([]*SimulationSetting, error)
	IsOperationReady(ctx context.Context, op *types.Operation) (*EndorserResult, error)
	DependencyState(ctx context.Context, result *EndorserResult) (*EndorserResultState, error)
	ConstraintsMet(ctx context.Context, result *EndorserResult) (bool, error)
}
