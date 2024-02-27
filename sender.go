package bundler

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/0xsequence/bundler/contracts/gen/solabis/abivalidator"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/types"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/ethwallet"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi/bind"
	"github.com/0xsequence/ethkit/go-ethereum/common"

	ethtypes "github.com/0xsequence/ethkit/go-ethereum/core/types"
)

type Sender struct {
	ID uint32

	Wallet    *ethwallet.Wallet
	Mempool   *Mempool
	Collector *Collector

	Provider *ethrpc.Provider
	ChainID  *big.Int

	executor *abivalidator.OperationValidator
}

func NewSender(id uint32, wallet *ethwallet.Wallet, mempool *Mempool, provider *ethrpc.Provider, executor *abivalidator.OperationValidator, collector *Collector) *Sender {
	chainID, err := provider.ChainID(context.TODO())
	if err != nil {
		panic(err)
	}

	return &Sender{
		ID:        id,
		Wallet:    wallet,
		Mempool:   mempool,
		Collector: collector,

		Provider: provider,
		ChainID:  chainID,

		executor: executor,
	}
}

func (s *Sender) Run(ctx context.Context) {
	for ctx.Err() == nil {
		ops := s.Mempool.ReserveOps(ctx, func(to []*TrackedOperation) []*TrackedOperation {
			if len(to) != 0 {
				return []*TrackedOperation{to[0]}
			}

			return nil
		})

		if len(ops) == 0 {
			time.Sleep(time.Second)
			continue
		}

		// Random delay between 0 and 1 second
		// it reduces the chances to collide with other senders
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		op := ops[0]
		paid, lied, err := s.simulateOperation(ctx, &op.Operation)

		// If we got an error, we should discard the operation
		if err != nil {
			s.Mempool.logger.Warn("sender: error simulating operation", "op", op.Digest(), "error", err)
			s.Mempool.DiscardOps(ctx, []*TrackedOperation{op})
			continue
		}

		// If the endorser lied to us, we should discard the operation
		// TODO: We should ban the endorser too
		if !paid {
			if lied {
				s.Mempool.logger.Warn("sender: endorser lied", "op", op.Digest(), "endorser", op.Endorser)
			} else {
				s.Mempool.logger.Info("sender: stale operation", "op", op.Digest())
			}
			s.Mempool.DiscardOps(ctx, []*TrackedOperation{op})
			continue
		}

		var calldataGas int64
		for _, b := range op.Calldata {
			switch b {
			case 0:
				calldataGas += 4
			default:
				calldataGas += 16
			}
		}

		nonce, err := s.Wallet.GetNonce(ctx)
		if err != nil {
			s.Mempool.logger.Warn("sender: error signing transaction", "op", op.Digest(), "error", err)
			s.Mempool.ReleaseOps(ctx, []*TrackedOperation{op}, ReadyAtChangeNone)
			continue
		}

		priorityFeePerGas := s.Collector.MinPriorityFee()

		signedTx, err := s.executor.SafeExecute(
			&bind.TransactOpts{
				Signer: func(a common.Address, t *ethtypes.Transaction) (*ethtypes.Transaction, error) {
					return s.Wallet.SignTx(t, s.ChainID)
				},
				Nonce:     new(big.Int).SetUint64(nonce),
				GasTipCap: priorityFeePerGas,
				NoSend:    true,
			},
			op.Entrypoint,
			op.Calldata,
			op.GasLimit,
			op.MaxFeePerGas,
			op.PriorityFeePerGas,
			op.FeeToken,
			op.BaseFeeScalingFactor,
			op.BaseFeeNormalizationFactor,
			big.NewInt(calldataGas),
		)

		if err != nil {
			s.Mempool.logger.Warn("sender: error signing transaction", "op", op.Digest(), "error", err)
			s.Mempool.ReleaseOps(ctx, []*TrackedOperation{op}, ReadyAtChangeNone)
			continue
		}

		// Try sending the transaction
		_, wait, err := s.Wallet.SendTransaction(ctx, signedTx)
		if err != nil {
			s.Mempool.logger.Warn("sender: error sending transaction", "op", op.Digest(), "error", err)
			s.Mempool.ReleaseOps(ctx, []*TrackedOperation{op}, ReadyAtChangeNone)
			continue
		}

		receipt, err := wait(ctx)
		if err != nil {
			s.Mempool.logger.Warn("sender: error waiting for receipt", "op", op.Digest(), "error", err)
			s.Mempool.ReleaseOps(ctx, []*TrackedOperation{op}, ReadyAtChangeNone)
			continue
		}

		s.Mempool.logger.Info("sender: operation executed", "op", op.Digest(), "tx", receipt.TxHash.String())
		s.Mempool.ReleaseOps(ctx, []*TrackedOperation{op}, ReadyAtChangeZero)
	}
}

func (s *Sender) simulateOperation(ctx context.Context, op *types.Operation) (paid bool, lied bool, err error) {
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
		big.NewInt(0),
	)

	if err != nil {
		return false, false, fmt.Errorf("unable to call simulateOperation: %w", err)
	}

	// The operation is healthy, ready to be executed
	if result.Paid {
		return true, false, nil
	}

	// The endorser is telling us that the operation was not ready
	// to be executed, so it didn't lie to us
	if !result.Readiness {
		return false, false, nil
	}

	// The only chance for the endorser left is that
	// he is returning a non-met contraint
	constraintsOk, err := endorser.CheckDependencyConstraints2(ctx, result.Dependencies, s.Provider)
	if err != nil {
		return false, false, fmt.Errorf("unable to check dependency constraints: %w", err)
	}

	// So constraints are not met, so it didn't lie to us
	if !constraintsOk {
		return false, false, nil
	}

	// The endorser is telling us that the operation was ready
	// to be executed, constraints are met, but we didn't get paid
	// this means the endorser lied to us.
	return false, true, nil
}
