package provider

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/0xsequence/ethkit/go-ethereum"
	"github.com/0xsequence/ethkit/go-ethereum/common"
)

type Batched struct {
	*Extended

	Slots *SlotsFetcher
}

type SimpleCall struct {
	Address common.Address
	Data    []byte
}

func NewBatched(provider *Extended, maxLatency time.Duration) *Batched {
	return &Batched{
		Extended: provider,

		Slots: NewSlotsFetcher(provider, maxLatency),
	}
}

func (b *Batched) Run(ctx context.Context) error {
	return b.Slots.Run(ctx)
}

func (b *Batched) StorageAtBatch(ctx context.Context, address common.Address, slots [][32]byte) ([][32]byte, error) {
	res, err := b.Slots.StorageAt(address, slots)

	select {
	case <-ctx.Done():
		// TODO: Maybe cancel the jobs?
		return nil, fmt.Errorf("fetcher: context cancelled")
	case err := <-err:
		return nil, err
	case res := <-res:
		return res, nil
	}
}

// Source: contracts/src/tools/BatchCaller.huff
const BatchCallerProgram = "0x60003560205b80361461003f57803590602001803590602001818160003791600091600060006000935af1503d8252906020013d6000823e3d0190610005565b60003580920382f3"
const BatchCallerPlaceholder = "0xf67dB61Ea957e88f9702D169D50C2e579766e089"

func BatchCall(ctx context.Context, provider *Extended, calls []*SimpleCall, overrides OverrideArgs) ([][]byte, error) {
	// Generate the call data
	// encoding all slots one after the other
	// the program takes max_size(32):(address(32):size(32):calldata)[]

	totalSize := 0
	maxSize := uint64(0)
	for _, call := range calls {
		totalSize += len(call.Data)
		if uint64(len(call.Data)) > maxSize {
			maxSize = uint64(len(call.Data))
		}
	}

	calldata := make([]byte, 32+totalSize+(len(calls)*64))

	// Append the max size padded to 32 bytes
	binary.BigEndian.PutUint64(calldata[24:], maxSize)
	windex := 32
	for _, call := range calls {
		// Write the address, padded to 32 bytes
		copy(calldata[windex+12:], call.Address.Bytes())
		windex += 32

		// Write the data size, padded to 32 bytes
		binary.BigEndian.PutUint64(calldata[windex+24:], uint64(len(call.Data)))
		windex += 32

		// Write the data
		copy(calldata[windex:], call.Data)
		windex += len(call.Data)
	}

	if overrides == nil {
		overrides = make(OverrideArgs)
	}

	bcp := common.HexToAddress(BatchCallerPlaceholder)
	bc := BatchCallerProgram
	overrides[bcp] = &Override{
		Code: &bc,
	}

	res, err := provider.CallWithOverride(ctx, &ethereum.CallMsg{
		To:   &bcp,
		Data: calldata,
	}, overrides)
	if err != nil {
		return nil, fmt.Errorf("fetcher: call failed: %w", err)
	}

	// Decode the response
	// returns (size(32):returndata)[]
	results := make([][]byte, len(calls))

	rindex := 0
	for i := 0; i < len(calls); i++ {
		if len(res) < rindex+32 {
			return nil, fmt.Errorf("fetcher: unexpected response length")
		}

		size := binary.BigEndian.Uint64(res[rindex+24 : rindex+32])
		rindex += 32
		results[i] = make([]byte, size)

		if len(res) < rindex+int(size) {
			return nil, fmt.Errorf("fetcher: unexpected response length")
		}

		copy(results[i], res[rindex:rindex+int(size)])
		rindex += int(size)
	}

	return results, nil
}
