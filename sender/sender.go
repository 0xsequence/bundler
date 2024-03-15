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
	"github.com/0xsequence/bundler/contracts/gen/solabis/abierc20"
	"github.com/0xsequence/bundler/contracts/gen/solabis/abivalidator"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/pricefeed"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/registry"
	"github.com/0xsequence/bundler/types"
	"github.com/0xsequence/ethkit/ethtxn"
	"github.com/0xsequence/ethkit/go-ethereum"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi/bind"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	ethtypes "github.com/0xsequence/ethkit/go-ethereum/core/types"
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

	Wallet    WalletInterface
	Provider  Provider
	Validator ValidatorInterface
	Mempool   mempool.Interface
	Endorser  endorser.Interface
	Collector collector.Interface
	Registry  registry.Interface
}

var _ Interface = &Sender{}

func NewSender(
	cfg *config.SendersConfig,
	logger *slog.Logger,
	id uint32,
	wallet WalletInterface,
	provider Provider,
	mempool mempool.Interface,
	endorser endorser.Interface,
	validator ValidatorInterface,
	Collector collector.Interface,
	Registry registry.Interface,
) *Sender {
	var chillWait time.Duration
	if cfg.ChillWait > 0 {
		chillWait = time.Duration(cfg.ChillWait) * time.Second
	} else {
		chillWait = 1 * time.Second
		logger.Warn("sender: chill wait not set, using default", "chillWait", chillWait)
	}

	var sleepWait time.Duration
	if cfg.SleepWait > 0 {
		sleepWait = time.Duration(cfg.SleepWait) * time.Millisecond
	} else {
		sleepWait = 100 * time.Millisecond
		logger.Warn("sender: sleep wait not set, using default", "sleepWait", sleepWait)
	}

	return &Sender{
		ID:     id,
		logger: logger,

		priorityFee: big.NewInt(int64(cfg.PriorityFee)),
		randomWait:  cfg.RandomWait,
		sleepWait:   sleepWait,
		chillWait:   chillWait,

		chilledOps: make(map[string]time.Time),
		blockedOps: make(map[string]struct{}),

		Wallet:    wallet,
		Provider:  provider,
		Mempool:   mempool,
		Endorser:  endorser,
		Validator: validator,
		Collector: Collector,
		Registry:  Registry,
	}
}

func (s *Sender) Run(ctx context.Context) {
	for ctx.Err() == nil {
		if !s.onRun(ctx) {
			time.Sleep(s.sleepWait)
		}
	}
}

func (s *Sender) onRun(ctx context.Context) bool {
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
		return false
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
		return true
	}

	// If the endorser lied to us, we should discard the operation
	if !res.Paid {
		if res.Lied {
			s.Registry.BanEndorser(op.Endorser, registry.PermanentBan)
			s.logger.Warn("sender: endorser lied", "op", opDigest, "endorser", op.Endorser, "innerOk", res.Meta.InnerOk, "innerPaid", res.Meta.InnerPaid.String(), "innerExpected", res.Meta.InnerExpected.String())
		} else {
			s.logger.Info("sender: stale operation", "op", opDigest)
		}
		s.Mempool.DiscardOps(ctx, []string{opDigest})
		return true
	}

	// Estimate the fixed calldata cost of the operation
	// this can be done by doing an estimate gas call to any other
	// address that is not a contract
	estimateAddr := common.HexToAddress("0x586FA0B5145FB12956dAaBD3b832Cc532d59230a")
	calldataGasLimit, err := s.Provider.EstimateGas(ctx, ethereum.CallMsg{
		To:   &estimateAddr,
		Data: op.Data,
	})
	if err != nil {
		s.logger.Warn("sender: error estimating gas", "op", opDigest, "error", err)
		s.Mempool.DiscardOps(ctx, []string{opDigest})
		return true
	}

	// See if it is profitable to execute the operation, for this we need to compute
	// the maximum gas price that we may pay for the operation
	// and compare it with the payment we are going to receive
	baseFee := s.Collector.BaseFee()
	native, priceSnap := s.Collector.NativeFeesPerGas(&op.Operation)
	effectiveFeePerGas := new(big.Int).Add(baseFee, native.MaxPriorityFeePerGas)
	if effectiveFeePerGas.Cmp(native.MaxFeePerGas) > 0 {
		effectiveFeePerGas.Set(native.MaxFeePerGas)
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
	if ourPayment.Cmp(payment) > 0 {
		s.logger.Info("sender: operation not profitable", "op", opDigest, "payment", payment.String(), "ourPayment", ourPayment.String())
		s.chilledOps[opDigest] = time.Now()
		s.Mempool.ReleaseOps(ctx, []string{opDigest}, proto.ReadyAtChange_None)
		return true
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
		return true
	}

	// Try sending the transaction
	_, wait, err := s.Wallet.SendTransaction(ctx, signedTx)
	if err != nil {
		s.logger.Warn("sender: error sending transaction", "op", opDigest, "error", err)
		s.Mempool.ReleaseOps(ctx, []string{opDigest}, proto.ReadyAtChange_None)
		return true
	}

	receipt, err := wait(ctx)
	if err != nil {
		s.logger.Warn("sender: error waiting for receipt", "op", opDigest, "error", err)
		s.Mempool.ReleaseOps(ctx, []string{opDigest}, proto.ReadyAtChange_None)
		return true
	}

	// Now that we have the receipt, we fire and forget the inspection
	go s.inspectReceipt(ctx, &op.Operation, receipt, priceSnap)

	s.logger.Info("sender: operation executed", "op", opDigest, "tx", receipt.TxHash.String())

	// Block the operation so we don't try to execute it again
	s.blockedOps[opDigest] = struct{}{}

	s.Mempool.ReleaseOps(ctx, []string{opDigest}, proto.ReadyAtChange_Zero)
	return true
}

func (s *Sender) IsChilled(op *types.Operation) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, ok := s.chilledOps[op.Hash()]
	fmt.Println("IsChilled?", op.Hash(), ok)
	return ok
}

func (s *Sender) IsBlocked(op *types.Operation) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, ok := s.blockedOps[op.Hash()]
	return ok
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

func (s *Sender) isOperationReady(
	ctx context.Context,
	op *types.Operation,
) (bool, error) {
	// We can check if the endorser is still returning `isValid == true`
	// since that would be a clear lie
	res, err := s.Endorser.IsOperationReady(ctx, op)
	if err != nil {
		return false, err
	}

	if !res.Readiness {
		return false, nil
	}

	ok, err := s.Endorser.ConstraintsMet(ctx, res)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (s *Sender) inspectReceipt(
	ctx context.Context,
	op *types.Operation,
	receipt *ethtypes.Receipt,
	priceSnap *pricefeed.Snapshot,
) {
	// If the transaction wasn't successful, two things may have happened:
	// - the operation was executed by someone else
	// - the endorser "lied" to us, and the simulation was wrong
	if receipt.Status == 0 {
		isReady, err := s.isOperationReady(ctx, op)
		if err != nil || !isReady {
			s.logger.Warn("inspector: likely operation collision", "op", op.Hash(), "tx", receipt.TxHash.String())
			// The operation was executed by someone else
			return
		}

		// The endorser lied to us
		// it is still marking the operation as ready
		// but the operation failed to execute
		s.Registry.BanEndorser(op.Endorser, registry.PermanentBan)
		s.logger.Error("inspector: endorser lied", "op", op.Hash(), "tx", receipt.TxHash.String())
		return
	}

	// If the operation was successful, we should check if we got paid
	// there are 3 possible outcomes:
	// - we got paid the expected amount or more
	// - we got paid less than expected
	// - we didn't get paid at all

	// For this check, we exploit the fact that the sender only sends a transaction
	// on each block, so we can check the balance before and after the transaction
	txBlockNum := receipt.BlockNumber

	prevBlockNum := new(big.Int).Sub(txBlockNum, big.NewInt(1))

	prevBalance, err := s.balanceOf(ctx, op.FeeToken, prevBlockNum)
	if err != nil {
		// We can't check the balance, so we can't do anything
		s.logger.Warn("inspector: unable to check prev balance", "op", op.Hash(), "tx", receipt.TxHash.String(), "error", err)
		return
	}
	nextBalance, err := s.balanceOf(ctx, op.FeeToken, txBlockNum)
	if err != nil {
		// We can't check the balance, so we can't do anything
		s.logger.Warn("inspector: unable to check next balance", "op", op.Hash(), "tx", receipt.TxHash.String(), "error", err)
		return
	}

	if receipt.EffectiveGasPrice == nil {
		s.logger.Warn("inspector: unable to check effective gas price", "op", op.Hash(), "tx", receipt.TxHash.String())
		return
	}

	balanceDiff := new(big.Int).Sub(nextBalance, prevBalance)
	nativeUsed := new(big.Int).Mul(receipt.EffectiveGasPrice, big.NewInt(int64(receipt.GasUsed)))

	isNative := op.FeeToken == common.Address{}
	if isNative {
		if balanceDiff.Sign() == 1 {
			// We got paid, end of story
			s.logger.Info(
				"inspector: operation paid",
				"op", op.Hash(),
				"tx", receipt.TxHash.String(),
				"amount", balanceDiff.String(),
			)
			return
		}
		s.logger.Warn(
			"inspector: operation not paid enough",
			"op", op.Hash(),
			"tx", receipt.TxHash.String(),
			"amount", balanceDiff.String(),
		)
	} else {
		// This is a bit more complicated, since we need to convert
		// the balanceDiff to native token and compare it with the nativeUsed
		nativePaid := priceSnap.ToNative(balanceDiff)
		if nativePaid.Cmp(nativeUsed) >= 0 {
			// We got paid, end of story
			s.logger.Info(
				"inspector: operation paid",
				"op", op.Hash(),
				"tx", receipt.TxHash.String(),
				"token", op.FeeToken,
				"amount", balanceDiff.String(),
			)
			return
		}
		s.logger.Warn(
			"inspector: operation not paid enough",
			"op", op.Hash(),
			"tx", receipt.TxHash.String(),
			"token", op.FeeToken,
			"amount", balanceDiff.String(),
			"nativePaid", nativePaid.String(),
			"nativeUsed", nativeUsed.String(),
		)
	}

	// The endorser lied to us
	s.Registry.BanEndorser(op.Endorser, registry.PermanentBan)
}

func (s *Sender) balanceOf(ctx context.Context, token common.Address, blockNum *big.Int) (*big.Int, error) {
	isNative := token == common.Address{}
	if isNative {
		return s.Provider.BalanceAt(ctx, s.Wallet.Address(), blockNum)
	}

	// Fetch from ERC20 balanceOf
	tokenContract, err := abierc20.NewERC20Caller(token, s.Provider)
	if err != nil {
		return nil, fmt.Errorf("unable to create ERC20Caller: %w", err)
	}

	return tokenContract.BalanceOf(&bind.CallOpts{
		BlockNumber: blockNum,
	}, s.Wallet.Address())
}
