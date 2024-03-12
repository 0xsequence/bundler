package pricefeed

import "math/big"

type Snapshot struct {
	ScalingFactor       *big.Int
	NormalizationFactor *big.Int
}

func (s *Snapshot) FromNative(native *big.Int) *big.Int {
	return new(big.Int).Div(new(big.Int).Mul(native, s.ScalingFactor), s.NormalizationFactor)
}

func (s *Snapshot) ToNative(value *big.Int) *big.Int {
	return new(big.Int).Div(new(big.Int).Mul(value, s.NormalizationFactor), s.ScalingFactor)
}
