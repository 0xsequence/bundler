package pricefeed

import (
	"context"
	"math/big"

	"github.com/0xsequence/ethkit/go-ethereum/common"
)

type NativeFeed struct{}

func (NativeFeed) Ready() bool {
	return true
}

func (NativeFeed) Name() string {
	return "native"
}

func (NativeFeed) FromNative(amount *big.Int) (*big.Int, error) {
	return new(big.Int).Set(amount), nil
}

func (NativeFeed) ToNative(amount *big.Int) (*big.Int, error) {
	return new(big.Int).Set(amount), nil
}

func (NativeFeed) Factors() (*big.Int, *big.Int, error) {
	return common.Big1, common.Big1, nil
}

func (NativeFeed) Start(ctx context.Context) error {
	return nil
}
