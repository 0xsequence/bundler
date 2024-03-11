package sender

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"math/rand"
	"sort"
	"time"

	"github.com/0xsequence/bundler/calldata"
	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/contracts/gen/solabis/abivalidator"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/types"
	"github.com/0xsequence/ethkit/ethtxn"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi/bind"
)

type Sender struct {
	ID uint32

	logger      *slog.Logger
	priorityFee *big.Int
	randomWait  int
	sleepWait   time.Duration

	Wallet        WalletInterface
	Validator     ValidatorInterface
	Mempool       mempool.Interface
	Endorser      endorser.Interface
	CalldataModel calldata.CostModel
}

var _ Interface = &Sender{}

func NewSender(
	cfg *config.SendersConfig,
	logger *slog.Logger,
	id uint32,
	wallet WalletInterface,
	mempool mempool.Interface,
	endorser endorser.Interface,
	validator ValidatorInterface,
	calldataModel calldata.CostModel,
) *Sender {
	return &Sender{
		ID:     id,
		logger: logger,

		priorityFee: big.NewInt(int64(cfg.PriorityFee)),
		randomWait:  cfg.RandomWait,
		sleepWait:   time.Duration(cfg.SleepWait),

		Wallet:        wallet,
		Mempool:       mempool,
		Endorser:      endorser,
		Validator:     validator,
		CalldataModel: calldataModel,
	}
}

func (s *Sender) Run(ctx context.Context) {
	for ctx.Err() == nil {
		ops := s.Mempool.ReserveOps(ctx, func(to []*mempool.TrackedOperation) []*mempool.TrackedOperation {
			if len(to) == 0 {
				return nil
			}

			// Sort them by highest value and pick the first one
			sort.Slice(to, func(i, j int) bool {
				return to[i].Value(s.CalldataModel).Cmp(to[j].Value(s.CalldataModel)) > 0
			})

			return to[:1]
		})

		if len(ops) == 0 {
			time.Sleep(s.sleepWait * time.Millisecond)
			continue
		}

		// Random delay between 0 and 1 second
		// it reduces the chances to collide with other senders
		if s.randomWait > 0 {
			time.Sleep(time.Duration(rand.Intn(s.randomWait)) * time.Millisecond)
		}

		op := ops[0]
		opDigest := ops[0].Hash()
		res, err := s.simulateOperation(ctx, &op.Operation)

		// If we got an error, we should discard the operation
		if err != nil {
			s.logger.Warn("sender: error simulating operation", "op", opDigest, "error", err)
			s.Mempool.DiscardOps(ctx, []string{opDigest})
			continue
		}

		// If the endorser lied to us, we should discard the operation
		// TODO: We should ban the endorser too
		if !res.Paid {
			if res.Lied {
				s.logger.Warn("sender: endorser lied", "op", opDigest, "endorser", op.Endorser, "innerOk", res.Meta.InnerOk, "innerPaid", res.Meta.InnerPaid.String(), "innerExpected", res.Meta.InnerExpected.String())
			} else {
				s.logger.Info("sender: stale operation", "op", opDigest)
			}
			s.Mempool.DiscardOps(ctx, []string{opDigest})
			continue
		}

		// Add the calldataGasLimit to the gasLimit of the op
		cgl := s.CalldataModel.CostFor(op.Data)

		signedTx, err := s.Wallet.NewTransaction(
			ctx,
			&ethtxn.TransactionRequest{
				GasPrice: op.MaxFeePerGas,
				GasTip:   s.priorityFee,
				GasLimit: op.GasLimit.Uint64() + cgl,
				To:       &op.Entrypoint,
				ETHValue: big.NewInt(0),
				Data:     op.Data,
			},
		)

		if err != nil {
			s.logger.Warn("sender: error signing transaction", "op", opDigest, "error", err)
			s.Mempool.ReleaseOps(ctx, []string{opDigest}, proto.ReadyAtChange_None)
			continue
		}

		// Try sending the transaction
		_, wait, err := s.Wallet.SendTransaction(ctx, signedTx)
		if err != nil {
			s.logger.Warn("sender: error sending transaction", "op", opDigest, "error", err)
			s.Mempool.ReleaseOps(ctx, []string{opDigest}, proto.ReadyAtChange_None)
			continue
		}

		receipt, err := wait(ctx)
		if err != nil {
			s.logger.Warn("sender: error waiting for receipt", "op", opDigest, "error", err)
			s.Mempool.ReleaseOps(ctx, []string{opDigest}, proto.ReadyAtChange_None)
			continue
		}

		s.logger.Info("sender: operation executed", "op", opDigest, "tx", receipt.TxHash.String())
		s.Mempool.ReleaseOps(ctx, []string{opDigest}, proto.ReadyAtChange_Zero)
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
	result, err := s.Validator.SimulateOperation(
		&bind.CallOpts{
			Context: ctx,
			From:    s.Wallet.Address(),
		},
		op.Endorser,
		*endorser.ToExecutorInput(&op.IEndorserOperation),
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
