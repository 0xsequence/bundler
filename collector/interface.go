package collector

import (
	"errors"
	"math/big"

	"github.com/0xsequence/bundler/pricefeed"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/types"
)

type Interface interface {
	BaseFee() *big.Int
	PriorityFee() *big.Int
	Feeds() []pricefeed.Feed
	ValidatePayment(op *types.Operation) error
	FeeAsks() (*proto.FeeAsks, error)
}

var InsufficientFeeError = errors.New("insufficient fee")
