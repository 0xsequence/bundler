package calldata

type CostModel interface {
	CostFor(calldata []byte) uint64
}
