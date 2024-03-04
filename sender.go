package bundler

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"math/rand"
	"time"

	"github.com/0xsequence/bundler/collector"
	"github.com/0xsequence/bundler/contracts/gen/solabis/abivalidator"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/types"
	"github.com/0xsequence/ethkit/ethwallet"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi/bind"
	"github.com/0xsequence/ethkit/go-ethereum/common"

	ethtypes "github.com/0xsequence/ethkit/go-ethereum/core/types"
)

type Sender struct {
	ID uint32

	logger *slog.Logger

	Wallet    *ethwallet.Wallet
	Mempool   mempool.Interface
	Collector *collector.Collector

	Endorser endorser.Interface
	ChainID  *big.Int

	executor *abivalidator.OperationValidator
}

func NewSender(logger *slog.Logger, id uint32, wallet *ethwallet.Wallet, mempool mempool.Interface, endorser endorser.Interface, executor *abivalidator.OperationValidator, collector *collector.Collector, chainID *big.Int) *Sender {
	return &Sender{
		ID:        id,
		logger:    logger,
		Wallet:    wallet,
		Mempool:   mempool,
		Collector: collector,
		Endorser:  endorser,

		ChainID: chainID,

		executor: executor,
	}
}

func (s *Sender) Run(ctx context.Context) {
	for ctx.Err() == nil {
		ops := s.Mempool.ReserveOps(ctx, func(to []*mempool.TrackedOperation) []*mempool.TrackedOperation {
			if len(to) != 0 {
				return []*mempool.TrackedOperation{to[0]}
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
		res, err := s.simulateOperation(ctx, &op.Operation)

		// If we got an error, we should discard the operation
		if err != nil {
			s.logger.Warn("sender: error simulating operation", "op", op.Digest(), "error", err)
			s.Mempool.DiscardOps(ctx, []*mempool.TrackedOperation{op})
			continue
		}

		// If the endorser lied to us, we should discard the operation
		// TODO: We should ban the endorser too
		if !res.Paid {
			if res.Lied {
				s.logger.Warn("sender: endorser lied", "op", op.Digest(), "endorser", op.Endorser, "innerOk", res.Meta.InnerOk, "innerPaid", res.Meta.InnerPaid.String(), "innerExpected", res.Meta.InnerExpected.String())
			} else {
				s.logger.Info("sender: stale operation", "op", op.Digest())
			}
			s.Mempool.DiscardOps(ctx, []*mempool.TrackedOperation{op})
			continue
		}

		nonce, err := s.Wallet.GetNonce(ctx)
		if err != nil {
			s.logger.Warn("sender: error signing transaction", "op", op.Digest(), "error", err)
			s.Mempool.ReleaseOps(ctx, []*mempool.TrackedOperation{op}, mempool.ReadyAtChangeNone)
			continue
		}

		priorityFeePerGas := s.Collector.PriorityFee()

		signedTx, err := s.executor.SafeExecute(
			&bind.TransactOpts{
				Signer: func(a common.Address, t *ethtypes.Transaction) (*ethtypes.Transaction, error) {
					return s.Wallet.SignTx(t, s.ChainID)
				},
				Nonce:     new(big.Int).SetUint64(nonce),
				GasTipCap: priorityFeePerGas,
				NoSend:    true,
				From:      s.Wallet.Address(),
			},
			op.Entrypoint,
			op.Calldata,
			op.GasLimit,
			op.MaxFeePerGas,
			op.PriorityFeePerGas,
			op.FeeToken,
			op.BaseFeeScalingFactor,
			op.BaseFeeNormalizationFactor,
		)

		if err != nil {
			s.logger.Warn("sender: error signing transaction", "op", op.Digest(), "error", err)
			s.Mempool.ReleaseOps(ctx, []*mempool.TrackedOperation{op}, mempool.ReadyAtChangeNone)
			continue
		}

		// Try sending the transaction
		_, wait, err := s.Wallet.SendTransaction(ctx, signedTx)
		if err != nil {
			s.logger.Warn("sender: error sending transaction", "op", op.Digest(), "error", err)
			s.Mempool.ReleaseOps(ctx, []*mempool.TrackedOperation{op}, mempool.ReadyAtChangeNone)
			continue
		}

		receipt, err := wait(ctx)
		if err != nil {
			s.logger.Warn("sender: error waiting for receipt", "op", op.Digest(), "error", err)
			s.Mempool.ReleaseOps(ctx, []*mempool.TrackedOperation{op}, mempool.ReadyAtChangeNone)
			continue
		}

		s.logger.Info("sender: operation executed", "op", op.Digest(), "tx", receipt.TxHash.String())
		s.Mempool.ReleaseOps(ctx, []*mempool.TrackedOperation{op}, mempool.ReadyAtChangeZero)
	}
}

type SimulateResult struct {
	Paid bool
	Lied bool
	Meta *SimulateResultMeta
}

type SimulateResultMeta struct {
	InnerOk       bool
	InnerPaid     big.Int
	InnerExpected big.Int
}

func parseMeta(res *abivalidator.OperationValidatorSimulationResult) (*SimulateResultMeta, error) {
	if len(res.Err) != 32*3+4 {
		return nil, fmt.Errorf("invalid error length, expected 32*3, got %v", len(res.Err))
	}

	ok := new(big.Int).SetBytes(res.Err[4:32+4]).Cmp(big.NewInt(1)) == 0
	paid := new(big.Int).SetBytes(res.Err[32+4 : 64+4])
	expected := new(big.Int).SetBytes(res.Err[64+4 : 96+4])

	return &SimulateResultMeta{
		InnerOk:       ok,
		InnerPaid:     *paid,
		InnerExpected: *expected,
	}, nil
}

func (s *Sender) simulateOperation(ctx context.Context, op *types.Operation) (*SimulateResult, error) {
	result, err := s.executor.SimulateOperation(
		&bind.CallOpts{
			Context: ctx,
			From:    s.Wallet.Address(),
		},
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
	)

	if err != nil {
		return nil, fmt.Errorf("unable to call simulateOperation: %w", err)
	}

	// The operation is healthy, ready to be executed
	if result.Paid {
		return &SimulateResult{
			Paid: true,
		}, nil
	}

	// The endorser is telling us that the operation was not ready
	// to be executed, so it didn't lie to us
	if !result.Readiness {
		return &SimulateResult{
			Paid: false,
			Lied: false,
		}, nil
	}

	// The only chance for the endorser left is that
	// he is returning a non-met contraint

	constraintsOk, err := s.Endorser.ConstraintsMet(ctx, endorser.FromExecutorResult(&result))
	if err != nil {
		return nil, fmt.Errorf("unable to check dependency constraints: %w", err)
	}

	// So constraints are not met, so it didn't lie to us
	if !constraintsOk {
		return &SimulateResult{
			Paid: false,
			Lied: false,
		}, nil
	}

	// The endorser is telling us that the operation was ready
	// to be executed, constraints are met, but we didn't get paid
	// this means the endorser lied to us.
	meta, err := parseMeta(&result)
	if err != nil {
		return nil, fmt.Errorf("unable to parse simulation meta: %w", err)
	}

	return &SimulateResult{
		Paid: false,
		Lied: true,
		Meta: meta,
	}, nil
}
