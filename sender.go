package bundler

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/0xsequence/bundler/contracts/gen/solabis/abivalidator"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/types"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/ethwallet"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi/bind"
	ethtypes "github.com/0xsequence/ethkit/go-ethereum/core/types"
	"github.com/davecgh/go-spew/spew"
)

const BatchSize = 1

type Sender struct {
	ID uint32

	Wallet   *ethwallet.Wallet
	Mempool  *Mempool
	Provider *ethrpc.Provider
	ChainID  *big.Int

	executor *abivalidator.OperationValidator
}

func NewSender(id uint32, wallet *ethwallet.Wallet, mempool *Mempool, provider *ethrpc.Provider, executor *abivalidator.OperationValidator) *Sender {
	chainID, err := provider.ChainID(context.TODO())
	if err != nil {
		panic(err)
	}

	return &Sender{
		ID:       id,
		Wallet:   wallet,
		Mempool:  mempool,
		Provider: provider,
		ChainID:  chainID,

		executor: executor,
	}
}

func (s *Sender) Run(ctx context.Context) {
	var execute, discard []*TrackedOperation

	for ctx.Err() == nil {
		ops := s.Mempool.ReserveOps(ctx, func(to []*TrackedOperation) []*TrackedOperation {
			if BatchSize < len(to) {
				return to[:BatchSize]
			} else {
				return to
			}
		})

		if len(ops) == 0 {
			time.Sleep(time.Second)
			continue
		}

		for _, op := range ops {
			result, err := endorser.IsOperationReady(ctx, s.Provider, &op.Operation)
			if err != nil {
			}
			if op.EndorserResult.Readiness {
				execute = append(execute, op)
			} else {
				discard = append(discard, op)
			}

			op.EndorserResult = result
		}

		s.Mempool.DiscardOps(ctx, discard)
		discard = nil

		for _, op := range execute {
			to := op.Entrypoint
			data := op.Calldata
			tx, err := s.Wallet.SignTx(ethtypes.NewTx(&ethtypes.DynamicFeeTx{
				To:   &to,
				Data: data,
			}), s.ChainID)
			if err != nil {
			}

			_, wait, err := s.Wallet.SendTransaction(ctx, tx)
			if err != nil {
			}

			receipt, err := wait(ctx)
			if err != nil {
			}

			spew.Dump(receipt)
		}

		execute = nil
	}
}

func (s *Sender) simulateOperation(ctx context.Context, op *types.Operation) (paid bool, lied bool, err error) {
	// TODO: compute this properly later
	var callDataGasUsage int64
	for _, b := range op.Calldata {
		switch b {
		case 0:
			callDataGasUsage += 4
		default:
			callDataGasUsage += 16
		}
	}

	result, err := s.executor.SimulateOperation(
		&bind.CallOpts{Context: ctx},
		op.Entrypoint,
		op.Calldata,
		op.EndorserCallData,
		op.GasLimit,
		op.MaxFeePerGas,
		op.PriorityFeePerGas,
		op.FeeToken,
		op.BaseFeeScalingFactor,
		op.BaseFeeNormalizationFactor,
		op.HasUntrustedContext,
		op.Endorser,
		big.NewInt(callDataGasUsage),
	)

	if err != nil {
		return false, false, fmt.Errorf("unable to call simulateOperation: %w", err)
	}

	// if result.lied is true and the constraints are all satisfied, then we should ban the endorser
	// if result.paid is false, then we should drop the transaction

	// TODO: only return lied if the constraints are all satisfied
	if result.Readiness {
		return false, true, nil
	}

	if result.Paid {
		return true, false, nil
	}

	return false, false, nil
}
