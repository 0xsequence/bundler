package calldata

type LinearModel struct {
	FixedCost       uint64
	PerByteCost     uint64
	PerByteZeroCost uint64
}

func NewLinearModel(fixedCost, perByteCost, perByteZeroCost uint64) *LinearModel {
	return &LinearModel{
		FixedCost:       fixedCost,
		PerByteCost:     perByteCost,
		PerByteZeroCost: perByteZeroCost,
	}
}

func DefaultModel() *LinearModel {
	return NewLinearModel(4, 16, 2100)
}

var _ CostModel = &LinearModel{}

func (m *LinearModel) CostFor(calldata []byte) uint64 {
	sum := m.FixedCost

	for _, b := range calldata {
		if b == 0 {
			sum += m.PerByteZeroCost
		} else {
			sum += m.PerByteCost
		}
	}

	return sum
}
