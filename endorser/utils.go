package endorser

import (
	"github.com/0xsequence/bundler/contracts/gen/solabis/abiendorser"
	"github.com/0xsequence/ethkit/go-ethereum/common"
)

func (e *EndorserResult) MustHaveDeps() {
	if e.Dependencies == nil {
		e.Dependencies = make([]abiendorser.EndorserDependency, 0)
	}
}

func (e *EndorserResult) UseDependency(address common.Address) *Dependency {
	e.MustHaveDeps()

	// Find the dependency
	for i := range e.Dependencies {
		dep := &e.Dependencies[i]
		if dep.Addr == address {
			return dep
		}
	}

	// Create a new dependency
	dep := abiendorser.EndorserDependency{
		Addr: address,
	}
	e.Dependencies = append(e.Dependencies, dep)

	return &dep
}

func (e *EndorserResult) UseGlobalDependency() *GlobalDependency {
	return &e.GlobalDependency
}

func (e *EndorserResult) SetBalance(address common.Address, enabled bool) {
	dep := e.UseDependency(address)
	dep.Balance = enabled
}

func (e *EndorserResult) SetCode(address common.Address, enabled bool) {
	dep := e.UseDependency(address)
	dep.Code = enabled
}

func (e *EndorserResult) SetStorageSlot(address common.Address, slot [32]byte, enabled bool) {
	dep := e.UseDependency(address)

	if len(dep.Slots) == 0 {
		dep.Slots = make([][32]byte, 0, 1)
	}

	dep.Slots = append(dep.Slots, slot)
}

func (e *EndorserResult) SetOrigin(enabled bool) {
	dep := e.UseGlobalDependency()
	dep.TxOrigin = enabled
}

func (e *EndorserResult) SetGasPrice(enabled bool) {
	dep := e.UseGlobalDependency()
	dep.TxGasPrice = enabled
}

func (e *EndorserResult) SetCoinbase(enabled bool) {
	dep := e.UseGlobalDependency()
	dep.Coinbase = enabled
}

func (e *EndorserResult) SetTimestamp(enabled bool) {
	dep := e.UseGlobalDependency()
	dep.Timestamp = enabled
}

func (e *EndorserResult) SetNumber(enabled bool) {
	dep := e.UseGlobalDependency()
	dep.Number = enabled
}

func (e *EndorserResult) SetDifficulty(enabled bool) {
	dep := e.UseGlobalDependency()
	dep.Difficulty = enabled
}

func (e *EndorserResult) SetChainID(enabled bool) {
	dep := e.UseGlobalDependency()
	dep.Chainid = enabled
}

func (e *EndorserResult) SetBasefee(enabled bool) {
	dep := e.UseGlobalDependency()
	dep.Basefee = enabled
}
