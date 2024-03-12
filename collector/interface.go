package collector

import (
	"errors"
	"math/big"

	"github.com/0xsequence/bundler/pricefeed"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/types"
)

type NativeFees struct {
	MaxFeePerGas         *big.Int
	MaxPriorityFeePerGas *big.Int
}

type Interface interface {
	BaseFee() *big.Int
	PriorityFee() *big.Int
	Cmp(a, b *types.Operation) int
	NativeFeesPerGas(a *types.Operation) (*NativeFees, *pricefeed.Snapshot)
	Feed(token string) (pricefeed.Feed, error)
	Feeds() []pricefeed.Feed
	ValidatePayment(op *types.Operation) error
	FeeAsks() (*proto.FeeAsks, error)
}

var InsufficientFeeError = errors.New("insufficient fee")
