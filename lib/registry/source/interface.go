package source

import (
	"math/big"

	"github.com/0xsequence/ethkit/go-ethereum/common"
)

type Interface interface {
	ReputationForEndorser(endorser common.Address) (*big.Int, error)
}
