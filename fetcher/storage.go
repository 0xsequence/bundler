package fetcher

import (
	"context"
	"fmt"

	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/common"
)

type CallOverride struct {
	Code string
}

// Source: contracts/src/tools/StorageFetcher.huff
const FetcherProgram = "0x60005b803554815260200136811061000257366000f3"

func FetchSlots(ctx context.Context, provider *ethrpc.Provider, address common.Address, slots [][32]byte) ([][32]byte, error) {
	// Generate the call data
	// encoding all slots one after the other
	calldata := make([]byte, 0, len(slots)*32)
	for _, slot := range slots {
		calldata = append(calldata, slot[:]...)
	}

	// Call the address, but set the code override to the fetcher program
	type Call struct {
		To   common.Address `json:"to"`
		Data string         `json:"data"`
	}

	estimateCall := &Call{
		To:   address,
		Data: "0x" + common.Bytes2Hex(calldata),
	}

	var res string
	rpcCall := ethrpc.NewCallBuilder[string]("eth_call", nil, estimateCall, nil, map[common.Address]*CallOverride{
		address: {Code: FetcherProgram},
	})
	_, err := provider.Do(ctx, rpcCall.Into(&res))
	if err != nil {
		return [][32]byte{}, err
	}

	resBytes := common.FromHex(res)
	if len(resBytes) != len(slots)*32 {
		return [][32]byte{}, fmt.Errorf("fetcher: unexpected response length")
	}

	// Decode the response
	results := make([][32]byte, len(slots))
	for i := 0; i < len(slots); i++ {
		copy(results[i][:], resBytes[i*32:(i+1)*32])
	}

	return results, nil
}
