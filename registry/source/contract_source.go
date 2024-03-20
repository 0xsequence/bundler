package source

import (
	"math/big"

	"github.com/0xsequence/bundler/contracts/gen/solabis/abiregistry"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi/bind"
	"github.com/0xsequence/ethkit/go-ethereum/common"
)

type ContractSource struct {
	Contract *abiregistry.RegistryCaller
}

func NewContractSource(provider bind.ContractCaller, addr common.Address) (*ContractSource, error) {
	contract, err := abiregistry.NewRegistryCaller(addr, provider)
	if err != nil {
		return nil, err
	}

	return &ContractSource{
		Contract: contract,
	}, nil
}

func (c *ContractSource) ReputationForEndorser(endorser common.Address) (*big.Int, error) {
	return c.Contract.Burn(nil, endorser)
}

var _ Interface = &ContractSource{}
