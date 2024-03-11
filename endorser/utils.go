package endorser

import (
	"math/big"

	"github.com/0xsequence/bundler/contracts/gen/solabis/abiendorser"
	"github.com/0xsequence/ethkit/go-ethereum/common"
)

func (e *EndorserResult) MustHaveDeps() {
	if e.Dependencies == nil {
		e.Dependencies = make([]abiendorser.IEndorserDependency, 0)
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
	dep := abiendorser.IEndorserDependency{
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
	dep.CoinBase = enabled
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
	dep.ChainId = enabled
}

func (e *EndorserResult) SetBasefee(enabled bool) {
	dep := e.UseGlobalDependency()
	dep.BaseFee = enabled
}

func (e1 *EndorserResult) Or(e2 *EndorserResult) *EndorserResult {
	if e1 == nil {
		return e2
	}

	if e2 == nil {
		return e1
	}

	e3 := &EndorserResult{}

	e3.WildcardOnly = e1.WildcardOnly || e2.WildcardOnly
	e3.Readiness = e1.Readiness && e2.Readiness
	e3.GlobalDependency = CombineGlobalDependency(e1.GlobalDependency, e2.GlobalDependency)
	e3.Dependencies = CombineDependencies(e1.Dependencies, e2.Dependencies)

	return e1
}

func CombineGlobalDependency(d1 GlobalDependency, d2 GlobalDependency) GlobalDependency {
	var maxBlockNumber, maxBlockTimestamp *big.Int

	if d1.MaxBlockNumber != nil && (d2.MaxBlockNumber == nil || d1.MaxBlockNumber.Cmp(d2.MaxBlockNumber) > 0) {
		maxBlockNumber = new(big.Int).Set(d1.MaxBlockNumber)
	} else if d2.MaxBlockNumber != nil {
		maxBlockNumber = new(big.Int).Set(d2.MaxBlockNumber)
	}

	if d1.MaxBlockTimestamp != nil && (d2.MaxBlockTimestamp == nil || d1.MaxBlockTimestamp.Cmp(d2.MaxBlockTimestamp) > 0) {
		maxBlockTimestamp = new(big.Int).Set(d1.MaxBlockTimestamp)
	} else if d2.MaxBlockTimestamp != nil {
		maxBlockTimestamp = new(big.Int).Set(d2.MaxBlockTimestamp)
	}

	return GlobalDependency{
		BaseFee:           d1.BaseFee || d2.BaseFee,
		BlobBaseFee:       d1.BlobBaseFee || d2.BlobBaseFee,
		ChainId:           d1.ChainId || d2.ChainId,
		CoinBase:          d1.CoinBase || d2.CoinBase,
		Difficulty:        d1.Difficulty || d2.Difficulty,
		GasLimit:          d1.GasLimit || d2.GasLimit,
		Number:            d1.Number || d2.Number,
		Timestamp:         d1.Timestamp || d2.Timestamp,
		TxOrigin:          d1.TxOrigin || d2.TxOrigin,
		TxGasPrice:        d1.TxGasPrice || d2.TxGasPrice,
		MaxBlockNumber:    maxBlockNumber,
		MaxBlockTimestamp: maxBlockTimestamp,
	}
}

func CombineDependencies(d1 []Dependency, d2 []Dependency) []Dependency {
	if d1 == nil {
		return d2
	}

	if d2 == nil {
		return d1
	}

	// Notice that some dependencies might be duplicated
	// in those cases, we need to combine the dependency itself
	m := make(map[common.Address]*Dependency, len(d1)+len(d2))
	for i := range d1 {
		dep := &d1[i]
		m[dep.Addr] = dep
	}

	for i := range d2 {
		if dep, ok := m[d2[i].Addr]; ok {
			m[d2[i].Addr] = CombineDependency(dep, &d2[i])
		} else {
			m[d2[i].Addr] = &d2[i]
		}
	}

	d3 := make([]Dependency, len(m))
	i := 0
	for _, dep := range m {
		d3[i] = *dep
		i++
	}

	return d3
}

func CombineDependency(d1 *Dependency, d2 *Dependency) *Dependency {
	if d1 == nil {
		return d2
	}

	if d2 == nil {
		return d1
	}

	d3 := &Dependency{
		Addr:    d1.Addr,
		Balance: d1.Balance || d2.Balance,
		Code:    d1.Code || d2.Code,
		Slots:   CombineSlots(d1.Slots, d2.Slots),
	}

	return d3
}

func CombineSlots(s1 [][32]byte, s2 [][32]byte) [][32]byte {
	if s1 == nil {
		return s2
	}

	if s2 == nil {
		return s1
	}

	// Some slots might be duplicated
	// in those cases, we need to combine the slots
	m := make(map[[32]byte]struct{}, len(s1)+len(s2))
	for _, s := range s1 {
		m[s] = struct{}{}
	}

	for _, s := range s2 {
		m[s] = struct{}{}
	}

	// Convert the map back to a slice
	s3 := make([][32]byte, len(m))
	i := 0
	for v := range m {
		s3[i] = v
		i++
	}

	return s3
}
