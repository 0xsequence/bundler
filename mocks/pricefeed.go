package mocks

import (
	"context"
	"math/big"

	"github.com/0xsequence/bundler/pricefeed"
	"github.com/stretchr/testify/mock"
)

type Feed struct {
	mock.Mock

	EtherPerUnit float64
	Decimals     uint
}

func (m *Feed) Factors() (*big.Int, *big.Int, error) {
	args := m.Called()
	return args.Get(0).(*big.Int), args.Get(1).(*big.Int), args.Error(2)
}

func (f *Feed) Ready() bool {
	return true
}

func (f *Feed) Name() string {
	return "mock"
}

func (f *Feed) FromNative(amount *big.Int) (*big.Int, error) {
	// amount / 1e18 / f.EtherPerUnit * 10 ^ f.Decimals

	numerator := new(big.Int).Mul(
		amount,
		new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(f.Decimals)), nil),
	)

	denominator := new(big.Int).Mul(
		big.NewInt(int64(f.EtherPerUnit*1000000000+0.5)),
		big.NewInt(1000000000),
	)

	return new(big.Int).Div(
		new(big.Int).Add(
			numerator,
			new(big.Int).Div(
				denominator,
				big.NewInt(2),
			),
		),
		denominator,
	), nil
}

func (f *Feed) ToNative(amount *big.Int) (*big.Int, error) {
	// amount / 10 ^ f.Decimals * f.EtherPerUnit * 1e18

	numerator := new(big.Int).Mul(
		amount,
		new(big.Int).Mul(
			big.NewInt(int64(f.EtherPerUnit*1000000000+0.5)),
			big.NewInt(1000000000),
		),
	)

	denominator := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(f.Decimals)), nil)

	return new(big.Int).Div(
		new(big.Int).Add(
			numerator,
			new(big.Int).Div(
				denominator,
				big.NewInt(2),
			),
		),
		denominator,
	), nil
}

func (f *Feed) Start(ctx context.Context) error {
	return nil
}

var _ pricefeed.Feed = &Feed{}
