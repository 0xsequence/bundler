package provider

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/0xsequence/ethkit/go-ethereum/common"
)

type FetchSlotJob struct {
	Address common.Address
	Slots   [][32]byte

	Res    chan [][32]byte
	ResErr chan error
}

func (j *FetchSlotJob) Complete(res [][32]byte, err error) {
	if err != nil {
		j.ResErr <- err
	} else {
		j.Res <- res
	}

	close(j.Res)
	close(j.ResErr)
}

func failJobs(js []*FetchSlotJob, err error) {
	for _, j := range js {
		j.Complete(nil, err)
	}
}

type SlotsFetcher struct {
	running *atomic.Int32

	Provider   *Extended
	MaxLatency time.Duration

	jobs chan FetchSlotJob
}

func NewSlotsFetcher(provider *Extended, maxLatency time.Duration) *SlotsFetcher {
	return &SlotsFetcher{
		running: &atomic.Int32{},

		MaxLatency: maxLatency,
		Provider:   provider,

		jobs: make(chan FetchSlotJob),
	}
}

func (sf *SlotsFetcher) Run(ctx context.Context) error {
	if !sf.running.CompareAndSwap(0, 1) {
		return fmt.Errorf("slots fetcher already running")
	}
	defer sf.running.Store(0)

	slotCount := 0
	buffer := make([]*FetchSlotJob, 0, 2048)
	timer := time.NewTimer(sf.MaxLatency)

	for {
		select {
		case job := <-sf.jobs:
			buffer = append(buffer, &job)
			slotCount += len(job.Slots)
			if slotCount >= 2048 {
				go sf.processBatch(ctx, buffer)
				buffer = make([]*FetchSlotJob, 0, 2048)
				slotCount = 0
				timer.Reset(sf.MaxLatency)
			}

		case <-timer.C:
			if slotCount > 0 {
				go sf.processBatch(ctx, buffer)

				buffer = make([]*FetchSlotJob, 0, 2048)
				slotCount = 0
			}

			timer.Reset(sf.MaxLatency)
		case <-ctx.Done():
			// Process any remaining jobs
			if slotCount > 0 {
				go sf.processBatch(ctx, buffer)
			}
			return nil
		}
	}
}

func (sf *SlotsFetcher) processBatch(ctx context.Context, jobs []*FetchSlotJob) {
	calls := make([]*SimpleCall, 0, len(jobs))
	override := make(OverrideArgs)

	for _, job := range jobs {
		calldata, overrideArgs := FetchSlotsEncode(job.Address, job.Slots)
		calls = append(calls, &SimpleCall{
			Address: job.Address,
			Data:    calldata,
		})

		err := override.Merge(overrideArgs)
		if err != nil {
			failJobs(jobs, err)
			return
		}
	}

	res, err := BatchCall(ctx, sf.Provider, calls, override)
	if err != nil {
		failJobs(jobs, fmt.Errorf("batch call failed: %w", err))
		return
	}

	for i, job := range jobs {
		decoded := res[i]
		results, err := FetchSlotsDecode(decoded)
		job.Complete(results, err)
	}
}

func (sf *SlotsFetcher) StorageAt(address common.Address, slots [][32]byte) (chan [][32]byte, chan error) {
	res := make(chan [][32]byte, 1)
	resErr := make(chan error, 1)

	sf.jobs <- FetchSlotJob{
		Address: address,
		Slots:   slots,
		Res:     res,
		ResErr:  resErr,
	}

	return res, resErr
}

// Source: contracts/src/tools/StorageFetcher.huff
const FetcherProgram = "0x60005b803554815260200136811061000257366000f3"

func FetchSlotsEncode(address common.Address, slots [][32]byte) ([]byte, OverrideArgs) {
	calldata := make([]byte, 0, len(slots)*32)
	for _, slot := range slots {
		calldata = append(calldata, slot[:]...)
	}

	fp := FetcherProgram
	overrideArgs := make(OverrideArgs)
	overrideArgs[address] = &Override{
		Code: &fp,
	}

	return calldata, overrideArgs
}

func FetchSlotsDecode(res []byte) ([][32]byte, error) {
	if len(res)%32 != 0 {
		return [][32]byte{}, fmt.Errorf("invalid response length")
	}

	// Decode the response
	size := len(res) / 32
	results := make([][32]byte, size)
	for i := 0; i < size; i++ {
		copy(results[i][:], res[i*32:(i+1)*32])
	}

	return results, nil
}
