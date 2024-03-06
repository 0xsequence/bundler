package partitioner

import (
	"sync"

	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/types"
	"golang.org/x/crypto/sha3"
)

type Partitioner struct {
	lock sync.Mutex

	OverlapLimit  uint
	WildcardLimit uint

	DependencyToOps  map[string][]*types.Operation
	OpToDependencies map[string][]string
}

func NewPartitioner(overlapLimit, wildcardLimit uint) *Partitioner {
	return &Partitioner{
		OverlapLimit:     overlapLimit,
		WildcardLimit:    wildcardLimit,
		DependencyToOps:  make(map[string][]*types.Operation),
		OpToDependencies: make(map[string][]string),
	}
}

const wildcard = "*"

func hash3(s0 []byte, s1 []byte, s2 []byte) string {
	h := sha3.New256()
	h.Write(s0)
	h.Write(s1)
	h.Write(s2)
	return string(h.Sum(nil))
}

func depHashes(dep *endorser.Dependency) []string {
	// AllSlots is a wildcard
	if dep.AllSlots {
		return []string{wildcard}
	}

	hs := make([]string, 0, len(dep.Slots)+3)

	for _, d := range dep.Slots {
		h := hash3([]byte{0}, dep.Addr[:], d[:])
		hs = append(hs, string(h))
	}

	if dep.Balance {
		h := hash3([]byte{1}, dep.Addr[:], []byte{1})
		hs = append(hs, string(h))
	}

	if dep.Code {
		h := hash3([]byte{1}, dep.Addr[:], []byte{2})
		hs = append(hs, string(h))
	}

	if dep.Nonce {
		h := hash3([]byte{1}, dep.Addr[:], []byte{3})
		hs = append(hs, string(h))
	}

	return hs
}

func depsOfResult(res *endorser.EndorserResult) []string {
	// Global dependencies are always a wildcard
	// TODO: Some of these may not be wildcards
	if res.WildcardOnly ||
		res.GlobalDependency.Basefee ||
		res.GlobalDependency.Blobbasefee ||
		res.GlobalDependency.Chainid ||
		res.GlobalDependency.Coinbase ||
		res.GlobalDependency.Difficulty ||
		res.GlobalDependency.GasLimit ||
		res.GlobalDependency.Number ||
		res.GlobalDependency.Timestamp ||
		res.GlobalDependency.TxOrigin ||
		res.GlobalDependency.TxGasPrice {
		return []string{wildcard}
	}

	hs := make([]string, 0, len(res.Dependencies)*2)

	for _, dep := range res.Dependencies {
		dhs := depHashes(&dep)
		if len(dhs) == 1 && dhs[0] == wildcard {
			return []string{wildcard}
		}

		hs = append(hs, dhs...)
	}

	return hs
}

func (p *Partitioner) Add(op *types.Operation, deps *endorser.EndorserResult) (bool, [][]*types.Operation) {
	oph := string(op.Hash())

	// Find all the dependencies that make the operation overlap
	// return a group for each overlap group, so that the caller
	// knows which operations may need to be removed to make room
	dhashes := depsOfResult(deps)

	p.lock.Lock()
	defer p.lock.Unlock()

	// If the operation is already known, this is a no-op
	// return true as it technically was added
	if _, ok := p.OpToDependencies[oph]; ok {
		return true, nil
	}

	if len(dhashes) != 0 && dhashes[0] == wildcard {
		if len(p.DependencyToOps[wildcard]) >= int(p.WildcardLimit) {
			// Copy for safety
			dh := dhashes[0]
			ops := make([]*types.Operation, len(p.DependencyToOps[dh]))
			copy(ops, p.DependencyToOps[dh])
			return false, [][]*types.Operation{ops}
		}
	} else {
		overlaps := make([][]*types.Operation, 0, len(dhashes))
		for _, dh := range dhashes {
			if len(p.DependencyToOps[dh]) >= int(p.OverlapLimit) {
				// Copy for safety
				ops := make([]*types.Operation, len(p.DependencyToOps[dh]))
				copy(ops, p.DependencyToOps[dh])
				overlaps = append(overlaps, ops)
			}
		}

		if len(overlaps) > 0 {
			return false, overlaps
		}
	}

	// Add the operation to the partitioner
	p.OpToDependencies[oph] = dhashes
	for _, dh := range dhashes {
		p.DependencyToOps[dh] = append(p.DependencyToOps[dh], op)
	}

	return true, nil
}

func (p *Partitioner) AddWildcard(op *types.Operation) (bool, [][]*types.Operation) {
	oph := string(op.Hash())

	p.lock.Lock()
	defer p.lock.Unlock()

	if _, ok := p.OpToDependencies[oph]; ok {
		return true, nil
	}

	if len(p.DependencyToOps[wildcard]) >= int(p.WildcardLimit) {
		// Copy for safety
		ops := make([]*types.Operation, len(p.DependencyToOps[wildcard]))
		copy(ops, p.DependencyToOps[wildcard])
		return false, [][]*types.Operation{ops}
	}

	// Add the operation to the partitioner
	p.OpToDependencies[oph] = []string{wildcard}
	p.DependencyToOps[wildcard] = append(p.DependencyToOps[wildcard], op)

	return true, nil
}

func (p *Partitioner) Remove(ops []string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, oph := range ops {
		// Remove the operation from the partitioner
		for _, dh := range p.OpToDependencies[oph] {
			ops := p.DependencyToOps[dh]
			for i, o := range ops {
				if o.Hash() == oph {
					ops = append(ops[:i], ops[i+1:]...)
					break
				}
			}
			p.DependencyToOps[dh] = ops
		}

		delete(p.OpToDependencies, oph)
	}
}
