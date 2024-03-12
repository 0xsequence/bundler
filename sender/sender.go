package sender

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/0xsequence/bundler/collector"
	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/contracts/gen/solabis/abivalidator"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/types"
	"github.com/0xsequence/ethkit/ethtxn"
	"github.com/0xsequence/ethkit/go-ethereum"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi/bind"
	"github.com/0xsequence/ethkit/go-ethereum/common"
)

type Sender struct {
	ID uint32

	logger      *slog.Logger
	priorityFee *big.Int
	randomWait  int
	sleepWait   time.Duration
	chillWait   time.Duration

	lock       sync.Mutex
	chilledOps map[string]time.Time
	blockedOps map[string]struct{}

	Wallet       WalletInterface
	GasEstimator GasEstimator
	Validator    ValidatorInterface
	Mempool      mempool.Interface
	Endorser     endorser.Interface
	Collector    collector.Interface
}

var _ Interface = &Sender{}

func NewSender(
	cfg *config.SendersConfig,
	logger *slog.Logger,
	id uint32,
	wallet WalletInterface,
	gasEstimator GasEstimator,
	mempool mempool.Interface,
	endorser endorser.Interface,
	validator ValidatorInterface,
	Collector collector.Interface,
) *Sender {
	return &Sender{
		ID:     id,
		logger: logger,

		priorityFee: big.NewInt(int64(cfg.PriorityFee)),
		randomWait:  cfg.RandomWait,
		sleepWait:   time.Duration(cfg.SleepWait),

		chilledOps: make(map[string]time.Time),
		blockedOps: make(map[string]struct{}),

		Wallet:       wallet,
		GasEstimator: gasEstimator,
		Mempool:      mempool,
		Endorser:     endorser,
		Validator:    validator,
		Collector:    Collector,
	}
}

func (s *Sender) Run(ctx context.Context) {
	for ctx.Err() == nil {
		s.onRun(ctx)
	}
}

func (s *Sender) onRun(ctx context.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Remove operations from chill list
	// if they already served their time
	for op, t := range s.chilledOps {
		if time.Since(t) > s.chillWait {
			delete(s.chilledOps, op)
		}
	}

	ops := s.Mempool.ReserveOps(ctx, func(to []*mempool.TrackedOperation) []*mempool.TrackedOperation {
		if len(to) == 0 {
			return nil
		}

		// Sort them by highest value and pick the first one
		sort.Slice(to, func(i, j int) bool {
			return s.Collector.Cmp(&to[i].Operation, &to[j].Operation) > 0
		})

		// Pick the first operation that is not blocked or chilled
		for _, op := range to {
			oph := op.Hash()
			if _, ok := s.blockedOps[oph]; ok {
				continue
			}
			if _, ok := s.chilledOps[oph]; ok {
				continue
			}
			return []*mempool.TrackedOperation{op}
		}

		// All operations are blocked or chilled
		// do not pick any of them
		return nil
	})

	if len(ops) == 0 {
		time.Sleep(s.sleepWait * time.Millisecond)
		return
	}

	// Random delay reduces the chances to collide with other senders
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
		return
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
		return
	}

	// Estimate the fixed calldata cost of the operation
	// this can be done by doing an estimate gas call to any other
	// address that is not a contract
	estimateAddr := common.HexToAddress("0x586FA0B5145FB12956dAaBD3b832Cc532d59230a")
	calldataGasLimit, err := s.GasEstimator.EstimateGas(ctx, ethereum.CallMsg{
		To:   &estimateAddr,
		Data: op.Data,
	})
	if err != nil {
		s.logger.Warn("sender: error estimating gas", "op", opDigest, "error", err)
		s.Mempool.DiscardOps(ctx, []string{opDigest})
		return
	}

	// See if it is profitable to execute the operation, for this we need to compute
	// the maximum gas price that we may pay for the operation
	// and compare it with the payment we are going to receive
	baseFee := s.Collector.BaseFee()
	maxFeePerGas, priorityFeePerGas := s.Collector.NativeFeesPerGas(&op.Operation)
	effectiveFeePerGas := new(big.Int).Add(baseFee, priorityFeePerGas)
	if effectiveFeePerGas.Cmp(maxFeePerGas) > 0 {
		effectiveFeePerGas.Set(maxFeePerGas)
	}

	// This defines what is the maximum fee per gas that we are going to pay
	ourMaxBaseFee := new(big.Int).Mul(baseFee, big.NewInt(2))
	ourMaxFeePerGas := new(big.Int).Add(ourMaxBaseFee, s.priorityFee)
	ourLikelyFeePerGas := new(big.Int).Add(baseFee, s.priorityFee)

	// The operation always pays fixedGas * effectiveFeePerGas
	payment := new(big.Int).Mul(big.NewInt(int64(calldataGasLimit)), effectiveFeePerGas)
	// We always need to pay the calldataGasLimit * likelyFeePerGas
	ourPayment := new(big.Int).Mul(big.NewInt(int64(calldataGasLimit)), ourLikelyFeePerGas)

	// The operation MAY X gasLimit, this is the same X that we need to pay, assume the total
	payment = payment.Add(payment, new(big.Int).Mul(op.GasLimit, effectiveFeePerGas))
	ourPayment = ourPayment.Add(ourPayment, new(big.Int).Mul(op.GasLimit, ourLikelyFeePerGas))

	// If we can't make profit from the operation, we should
	// chill it for a while. There are many ever-changing factors
	// that may make the operation profitable in the future
	if ourPayment.Cmp(payment) < 0 {
		s.logger.Info("sender: operation not profitable", "op", opDigest, "payment", payment.String(), "ourPayment", ourPayment.String())
		s.chilledOps[opDigest] = time.Now()
		s.Mempool.ReleaseOps(ctx, []string{opDigest}, proto.ReadyAtChange_None)
		return
	}

	signedTx, err := s.Wallet.NewTransaction(
		ctx,
		&ethtxn.TransactionRequest{
			GasPrice: ourMaxFeePerGas,
			GasTip:   s.priorityFee,
			GasLimit: calldataGasLimit + op.GasLimit.Uint64(),
			To:       &op.Entrypoint,
			ETHValue: big.NewInt(0),
			Data:     op.Data,
		},
	)

	if err != nil {
		s.logger.Warn("sender: error signing transaction", "op", opDigest, "error", err)
		s.Mempool.ReleaseOps(ctx, []string{opDigest}, proto.ReadyAtChange_None)
		return
	}

	// Try sending the transaction
	_, wait, err := s.Wallet.SendTransaction(ctx, signedTx)
	if err != nil {
		s.logger.Warn("sender: error sending transaction", "op", opDigest, "error", err)
		s.Mempool.ReleaseOps(ctx, []string{opDigest}, proto.ReadyAtChange_None)
		return
	}

	receipt, err := wait(ctx)
	if err != nil {
		s.logger.Warn("sender: error waiting for receipt", "op", opDigest, "error", err)
		s.Mempool.ReleaseOps(ctx, []string{opDigest}, proto.ReadyAtChange_None)
		return
	}

	s.logger.Info("sender: operation executed", "op", opDigest, "tx", receipt.TxHash.String())

	// Block the operation so we don't try to execute it again
	s.blockedOps[opDigest] = struct{}{}

	s.Mempool.ReleaseOps(ctx, []string{opDigest}, proto.ReadyAtChange_Zero)
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
	// he is returning a non-met constraint

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
