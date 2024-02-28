package collector

import (
	"math/big"

	"github.com/0xsequence/bundler/pricefeed"
	"github.com/0xsequence/bundler/types"
)

type Interface interface {
	BaseFee() *big.Int
	PriorityFee() *big.Int
	Feeds() []pricefeed.Feed
	MeetsPayment(op *types.Operation) (bool, error)
}
