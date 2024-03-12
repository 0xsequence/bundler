package partitioner_test

import (
	"testing"

	"github.com/0xsequence/bundler/contracts/gen/solabis/abiendorser"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/mempool/partitioner"
	"github.com/0xsequence/bundler/types"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestAdd1(t *testing.T) {
	p := partitioner.NewPartitioner(1, 1)

	ok, deps := p.Add(&types.Operation{}, &endorser.EndorserResult{})
	assert.True(t, ok)
	assert.Nil(t, deps)
}

func TestAdd2Independent(t *testing.T) {
	p := partitioner.NewPartitioner(2, 2)

	ok, deps := p.Add(&types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{2},
		},
	}, &endorser.EndorserResult{})
	assert.True(t, ok)
	assert.Nil(t, deps)

	ok, deps = p.Add(&types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{1},
		},
	}, &endorser.EndorserResult{})
	assert.True(t, ok)
	assert.Nil(t, deps)
}

func TestAdd2IndependentWithDependencies(t *testing.T) {
	p := partitioner.NewPartitioner(2, 2)

	ok, deps := p.Add(&types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{2},
		},
	}, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:    common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				Balance: true,
			},
		},
	})
	assert.True(t, ok)
	assert.Nil(t, deps)

	ok, deps = p.Add(&types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{1},
		},
	}, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:    common.HexToAddress("0x2e417D097fF04E4F532A7856a1b4c62a34988E16"),
				Balance: true,
			},
		},
	})
	assert.True(t, ok)
	assert.Nil(t, deps)
}

func TestAdd2IndependentWithDependenciesButRoom(t *testing.T) {
	p := partitioner.NewPartitioner(2, 2)

	ok, deps := p.Add(&types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{2},
		},
	}, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:    common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				Balance: true,
			},
		},
	})
	assert.True(t, ok)
	assert.Nil(t, deps)

	ok, deps = p.Add(&types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{1},
		},
	}, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:    common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				Balance: true,
			},
		},
	})
	assert.True(t, ok)
	assert.Nil(t, deps)
}

func TestBlockAddTwoWithDependencies(t *testing.T) {
	p := partitioner.NewPartitioner(1, 1)

	op1 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{2},
		},
	}

	op2 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{1},
		},
	}

	ok, deps := p.Add(op1, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:    common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				Balance: true,
			},
		},
	})
	assert.True(t, ok)
	assert.Nil(t, deps)

	ok, deps = p.Add(op2, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:    common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				Balance: true,
			},
		},
	})
	assert.False(t, ok)
	assert.Equal(t, deps, [][]*types.Operation{{op1}})
}

func TestBlockAddTwoWithWildcard(t *testing.T) {
	p := partitioner.NewPartitioner(1, 1)

	op1 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{2},
		},
	}

	op2 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{1},
		},
	}

	ok, deps := p.Add(op1, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:     common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				AllSlots: true,
			},
		},
	})
	assert.True(t, ok)
	assert.Nil(t, deps)

	ok, deps = p.AddWildcard(op2)
	assert.False(t, ok)
	assert.Equal(t, deps, [][]*types.Operation{{op1}})
}

func TestAddTwoWithWildcardWithRoom(t *testing.T) {
	p := partitioner.NewPartitioner(2, 2)

	op1 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{2},
		},
	}

	op2 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{1},
		},
	}

	ok, deps := p.Add(op1, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:     common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				AllSlots: true,
			},
		},
	})
	assert.True(t, ok)
	assert.Nil(t, deps)

	ok, deps = p.AddWildcard(op2)
	assert.True(t, ok)
	assert.Nil(t, deps)
}

func TestMultipleOverlap(t *testing.T) {
	p := partitioner.NewPartitioner(2, 2)

	op1 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{1},
		},
	}

	op2 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{2},
		},
	}

	op3 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{3},
		},
	}

	op4 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{4},
		},
	}

	ok, deps := p.Add(op1, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:  common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				Slots: [][32]byte{{1}},
			},
		},
	})

	assert.True(t, ok)
	assert.Nil(t, deps)

	ok, deps = p.Add(op2, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:  common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				Slots: [][32]byte{{1}},
			},
		},
	})

	assert.True(t, ok)
	assert.Nil(t, deps)

	ok, deps = p.Add(op3, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:  common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				Slots: [][32]byte{{2}},
			},
		},
	})

	assert.True(t, ok)
	assert.Nil(t, deps)

	ok, deps = p.Add(op4, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:  common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				Slots: [][32]byte{{2}, {1}},
			},
		},
	})

	assert.False(t, ok)
	assert.Equal(t, deps, [][]*types.Operation{{op1, op2}})
}

func TestMultidimentionalOverlap(t *testing.T) {
	p := partitioner.NewPartitioner(2, 2)

	op1 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{1},
		},
	}

	op2 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{2},
		},
	}

	op3 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{3},
		},
	}

	op4 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{4},
		},
	}

	op5 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{5},
		},
	}

	ok, deps := p.Add(op1, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:  common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				Slots: [][32]byte{{1}},
			},
		},
	})

	assert.True(t, ok)
	assert.Nil(t, deps)

	ok, deps = p.Add(op2, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:  common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				Slots: [][32]byte{{1}},
			},
		},
	})

	assert.True(t, ok)
	assert.Nil(t, deps)

	ok, deps = p.Add(op3, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:  common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				Slots: [][32]byte{{2}},
			},
		},
	})

	assert.True(t, ok)
	assert.Nil(t, deps)

	ok, deps = p.Add(op4, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:  common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				Slots: [][32]byte{{2}},
			},
		},
	})

	assert.True(t, ok)
	assert.Nil(t, deps)

	ok, deps = p.Add(op5, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:  common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				Slots: [][32]byte{{1}, {2}},
			},
		},
	})

	assert.False(t, ok)
	assert.Equal(t, deps, [][]*types.Operation{{op1, op2}, {op3, op4}})
}

func TestRemoveOps(t *testing.T) {
	p := partitioner.NewPartitioner(2, 2)

	op1 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{1},
		},
	}

	op2 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{2},
		},
	}

	op3 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{3},
		},
	}

	op4 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{4},
		},
	}

	op5 := &types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Data: []byte{5},
		},
	}

	ok, deps := p.Add(op1, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:  common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				Slots: [][32]byte{{1}},
			},
		},
	})

	assert.True(t, ok)
	assert.Nil(t, deps)

	ok, deps = p.Add(op2, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:  common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				Slots: [][32]byte{{1}},
			},
		},
	})

	assert.True(t, ok)
	assert.Nil(t, deps)

	ok, deps = p.Add(op3, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:  common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				Slots: [][32]byte{{2}},
			},
		},
	})

	assert.True(t, ok)
	assert.Nil(t, deps)

	ok, deps = p.Add(op4, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:  common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				Slots: [][32]byte{{2}},
			},
		},
	})

	assert.True(t, ok)
	assert.Nil(t, deps)

	p.Remove([]string{
		op1.Hash(),
		op4.Hash(),
	})

	ok, deps = p.Add(op5, &endorser.EndorserResult{
		Dependencies: []abiendorser.IEndorserDependency{
			{
				Addr:  common.HexToAddress("0x3B377376F325AbA4a5f2E5E3d143FD8cd15afCEd"),
				Slots: [][32]byte{{1}, {2}},
			},
		},
	})

	assert.True(t, ok)
	assert.Nil(t, deps)
}
