package pricefeed

import (
	"context"
	"math/big"
)

type NativeFeed struct{}

func (NativeFeed) Ready() bool {
	return true
}

func (NativeFeed) Name() string {
	return "native"
}

func (NativeFeed) Snapshot() (*Snapshot, error) {
	return &Snapshot{
		ScalingFactor:       big.NewInt(1),
		NormalizationFactor: big.NewInt(1),
	}, nil
}

func (NativeFeed) Start(ctx context.Context) error {
	return nil
}
