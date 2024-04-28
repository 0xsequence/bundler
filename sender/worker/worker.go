package worker

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/big"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/0xsequence/bundler/contracts/gen/solabis/abierc20"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/lib/collector"
	"github.com/0xsequence/bundler/lib/interfaces"
	"github.com/0xsequence/bundler/lib/pricefeed"
	"github.com/0xsequence/bundler/lib/registry"
	"github.com/0xsequence/bundler/lib/types"
	"github.com/0xsequence/bundler/lib/utils"
	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/ethkit/ethtxn"
	"github.com/0xsequence/ethkit/go-ethereum"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi/bind"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	ethtypes "github.com/0xsequence/ethkit/go-ethereum/core/types"
	"github.com/prometheus/client_golang/prometheus"
)

type ReleaseOp struct {
	Oph    string
	Change proto.ReadyAtChange
}

type BanEndorser struct {
	Endorser common.Address
	Type     registry.BanType
}

type OperationReady struct {
	Op *mempool.TrackedOperation
	Tx *ethtxn.TransactionRequest

	Price *pricefeed.Snapshot
}

type Worker struct {
	running atomic.Uint32

	metrics *metrics
	logger  *slog.Logger

	randomWait  int
	priorityFee *big.Int
	wallet      interfaces.Wallet
	minBalance  *big.Int

	ready   chan *OperationReady
	chill   chan string
	done    chan string
	discard chan string
	release chan *ReleaseOp
	ban     chan *BanEndorser
	pause   chan bool
	sus     chan *mempool.TrackedOperation

	Provider  interfaces.Provider
	Collector collector.Interface
	Endorser  endorser.Interface
	Simulator interfaces.Validator
}

func NewWorker(
	Provider interfaces.Provider,
	Collector collector.Interface,
	Endorser endorser.Interface,
	Simulator interfaces.Validator,
	Wallet interfaces.Wallet,
	PriorityFee *big.Int,
	MinBalance *big.Int,
) *Worker {
	return &Worker{
		running: atomic.Uint32{},

		metrics: createMetrics(),
		logger:  slog.Default(),

		priorityFee: PriorityFee,
		wallet:      Wallet,
		minBalance:  MinBalance,

		ready: make(chan *OperationReady),
		chill: make(chan string),
		done:  make(chan string),

		discard: make(chan string),
		release: make(chan *ReleaseOp),
		ban:     make(chan *BanEndorser),
		pause:   make(chan bool),

		sus: make(chan *mempool.TrackedOperation, 128),

		Provider:  Provider,
		Collector: Collector,
		Endorser:  Endorser,
		Simulator: Simulator,
	}
}

func (w *Worker) Chill() <-chan string {
	return w.chill
}

func (w *Worker) Done() <-chan string {
	return w.done
}

func (w *Worker) Discard() <-chan string {
	return w.discard
}

func (w *Worker) Release() <-chan *ReleaseOp {
	return w.release
}

func (w *Worker) Ban() <-chan *BanEndorser {
	return w.ban
}

func (w *Worker) Pause() {
	w.pause <- true
}

func (w *Worker) Resume() {
	w.pause <- false
}

func (w *Worker) SetRandomWait(wait int) {
	w.randomWait = wait
}

func (w *Worker) SetLogger(logger *slog.Logger) {
	w.logger = logger
}

func (w *Worker) SetRegisterer(reg prometheus.Registerer) {
	tagged := prometheus.WrapRegistererWith(prometheus.Labels{"sender": w.wallet.Address().String()}, reg)
	w.metrics.register(tagged)
}

func (w *Worker) Run(ctx context.Context, input <-chan *mempool.TrackedOperation) error {
	if !w.running.CompareAndSwap(0, 1) {
		return fmt.Errorf("worker: already running")
	}

	defer w.running.Store(0)
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	// Start the monitor
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.monitorWorker(ctx)
	}()

	// Start the preparer
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.prepareWorker(ctx, input)
	}()

	// Start the sus inspector
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.validateSusOperation(ctx)
	}()

	// Start the sender
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.sendWorker(ctx)
	}()

	w.logger.Info("worker: running", "wallet", w.wallet.Address().String())

	return nil
}

func (w *Worker) validateSusOperation(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case op := <-w.sus:
			if op != nil {
				// TODO: Do a proper inspection of the operations
				// it may be possible that the endorser lied to us
				// and the operation never intended to pay us.
				w.discard <- op.Hash()
			}
		}
	}
}

func (w *Worker) monitorWorker(ctx context.Context) {
	// Get the current balance of the wallet
	// if the balance is below the minimum, we pause
	// the worker until the balance is restored
	latch := false
	for ctx.Err() == nil {
		balance, err := w.Provider.BalanceAt(ctx, w.wallet.Address(), nil)
		if err != nil {
			w.logger.Warn("worker: error fetching balance", "error", err)
			continue
		}

		if balance.Cmp(w.minBalance) < 0 {
			if latch {
				w.logger.Warn("worker: balance below minimum, pausing worker", "balance", balance.String(), "minBalance", w.minBalance.String())
				w.Pause()
				latch = false
			}
		} else {
			if !latch {
				w.logger.Info("worker: balance restored, resuming worker", "balance", balance.String(), "minBalance", w.minBalance.String())
				w.Resume()
				latch = true
			}
		}

		// Wait 10 seconds
		// or until the context is done
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
		}
	}

}

func (w *Worker) prepareWorker(ctx context.Context, input <-chan *mempool.TrackedOperation) {
	var currentInput <-chan *mempool.TrackedOperation
	for {
		select {
		case <-ctx.Done():
			return
		case p := <-w.pause:
			if p {
				currentInput = nil
			} else {
				currentInput = input
			}
		case op := <-currentInput:
			if op != nil {
				w.doPrepare(ctx, op)
			}
		}
	}
}

func (w *Worker) sendWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case ready := <-w.ready:
			// Send operation
			w.doSend(ctx, ready)
		}
	}
}

func (w *Worker) staticGasLimit(ctx context.Context, op *mempool.TrackedOperation) (*big.Int, error) {
	// Estimate the fixed calldata cost of the operation
	// this can be done by doing an estimate gas call to any other
	// address that is not a contract
	estimateAddr := common.HexToAddress("0x586FA0B5145FB12956dAaBD3b832Cc532d59230a")
	calldataGasLimit, err := w.Provider.EstimateGas(ctx, ethereum.CallMsg{
		To:   &estimateAddr,
		Data: op.Data,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to estimate gas: %w", err)
	}

	return big.NewInt(int64(calldataGasLimit)), nil
}

func (w *Worker) doPrepare(ctx context.Context, op *mempool.TrackedOperation) {
	// Random delay reduces the chances to collide with other senders
	if w.randomWait > 0 {
		time.Sleep(time.Duration(rand.Intn(w.randomWait)) * time.Millisecond)
	}

	// Measure time after delay, delay is random
	defer utils.RecordFunctionDuration(time.Now(), w.metrics.prepareOpTime)

	staticGas, err := w.staticGasLimit(ctx, op)
	if err != nil {
		w.metrics.failedSendOps.With(w.metrics.failedEstimateGas).Inc()
		w.logger.Warn("sender: error estimating static gas usage", "op", op.Hash(), "error", err)

		w.release <- &ReleaseOp{Oph: op.Hash(), Change: proto.ReadyAtChange_None}
		return
	}

	opDigest := op.Hash()
	result, err := w.Simulator.SimulateOperation(
		&bind.CallOpts{
			Context: ctx,
			From:    w.wallet.Address(),
		},
		*endorser.ToSimulatorInput(&op.IEndorserOperation),
	)

	if err != nil {
		w.metrics.failedSendOps.With(w.metrics.failedSimulateOperation).Inc()
		w.logger.Warn("sender: error simulating operation", "op", opDigest, "error", err)
		w.release <- &ReleaseOp{Oph: opDigest, Change: proto.ReadyAtChange_None}
		return
	}

	// This is what it will cost to use to execute this operation
	// the cost is computed as (staticGas + result.gasUsed) * gasPrice
	// the gasPrice is the current baseFee + our priorityFee of choice
	baseFee := w.Collector.BaseFee()
	gasPrice := new(big.Int).Add(baseFee, w.priorityFee)
	cost := new(big.Int).Mul(new(big.Int).Add(staticGas, result.GasUsed), gasPrice)

	// Our cost is in native tokens, but the payment is in the operation's fee token
	// we need to convert the cost to the operation's fee token
	_, priceSnap := w.Collector.NativeFeesPerGas(&op.Operation)
	paymentNative := priceSnap.ToNative(result.Payment)

	// If the payment is below the cost, there is a chance the endorser lied
	// but it could also be that our priority fee (Or token price) is just different
	if cost.Cmp(paymentNative) > 0 {
		// Schedule for inspection
		w.logger.Warn("sender: operation payment below cost", "op", opDigest, "cost", cost.String(), "payment", paymentNative.String())

		// Attempt to put the operation into the `sus` channel
		// if full, then release the operation
		// we never want to block the prepare channel
		select {
		case w.sus <- op:
		default:
			w.release <- &ReleaseOp{Oph: opDigest, Change: proto.ReadyAtChange_None}
		}

		// TODO: Register on METRICS
		return
	}

	diffFloat, _ := new(big.Int).Sub(result.Payment, paymentNative).Float64()
	w.metrics.profitableOpDiff.Observe(diffFloat)

	opr := &OperationReady{
		Op:    op,
		Price: priceSnap,
		Tx: &ethtxn.TransactionRequest{
			GasPrice: gasPrice,
			GasTip:   w.priorityFee,
			GasLimit: staticGas.Add(staticGas, op.GasLimit).Uint64(),
			To:       &op.Entrypoint,
			ETHValue: big.NewInt(0),
			Data:     op.Data,
		},
	}

	// Attempt to forward the operation
	// if the ready channel is full, give it 100ms
	// if it still full, then return then release the operation
	// we do this because the operation may become invalid while
	// waiting for the ready channel
	select {
	case w.ready <- opr:
	case <-time.After(100 * time.Millisecond):
		w.metrics.preparedAndDroppedOps.Inc()
		w.release <- &ReleaseOp{Oph: opDigest, Change: proto.ReadyAtChange_Now}
	}
}

func (w *Worker) doSend(ctx context.Context, opr *OperationReady) {
	oph := opr.Op.Hash()

	// Always release the operation
	defer func() { w.release <- &ReleaseOp{Oph: oph, Change: proto.ReadyAtChange_None} }()
	defer utils.RecordFunctionDuration(time.Now(), w.metrics.sendOpTime)

	signedTx, err := w.wallet.NewTransaction(ctx, opr.Tx)

	if err != nil {
		w.metrics.failedSendOps.With(w.metrics.failedSignTransaction).Inc()
		w.logger.Warn("sender: error signing transaction", "op", oph, "error", err)
		return
	}

	// Try sending the transaction
	tx, wait, err := w.wallet.SendTransaction(ctx, signedTx)
	if err != nil {
		w.metrics.failedSendOps.With(w.metrics.failedSendTransaction).Inc()
		w.logger.Warn("sender: error sending transaction", "op", oph, "error", err)
		return
	}

	w.metrics.executedOps.Inc()

	startReceipt := time.Now()
	receipt, err := wait(ctx)
	if err != nil {
		w.metrics.failedReceiptOps.Inc()
		w.logger.Warn("sender: error waiting for receipt", "op", oph, "error", err)
		return
	}

	w.metrics.waitReceiptTime.Observe(time.Since(startReceipt).Seconds())

	// Now that we have the receipt, we fire and forget the inspection
	go w.inspectReceipt(ctx, &opr.Op.Operation, tx, receipt, opr.Price)

	w.logger.Info("sender: operation executed", "op", oph, "tx", receipt.TxHash.String())

	w.done <- oph
}

func (w *Worker) inspectReceipt(
	ctx context.Context,
	op *types.Operation,
	tx *ethtypes.Transaction,
	receipt *ethtypes.Receipt,
	priceSnap *pricefeed.Snapshot,
) {
	defer utils.RecordFunctionDuration(time.Now(), w.metrics.inspectReceiptTime)

	// If the transaction wasn't successful, two things may have happened:
	// - the operation was executed by someone else
	// - the endorser "lied" to us, and the simulation was wrong
	if receipt.Status == 0 {
		isReady, err := w.isOperationReady(ctx, op)
		if err != nil || !isReady {
			// The operation was executed by someone else
			w.metrics.inspectReceiptReverted.With(prometheus.Labels{"lied": "false"}).Inc()
			w.logger.Warn("inspector: likely operation collision", "op", op.Hash(), "tx", receipt.TxHash.String())
			return
		}

		// The endorser lied to us
		// it is still marking the operation as ready
		// but the operation failed to execute
		w.metrics.inspectReceiptReverted.With(prometheus.Labels{"lied": "true"}).Inc()
		w.logger.Error("inspector: endorser lied", "op", op.Hash(), "tx", receipt.TxHash.String())

		w.ban <- &BanEndorser{Endorser: op.Endorser, Type: registry.PermanentBan}
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
	prevBalance, err := w.balanceOf(ctx, op.FeeToken, prevBlockNum)
	if err != nil {
		// We can't check the balance, so we can't do anything
		w.metrics.inspectReceiptFailed.With(w.metrics.failedInspectReceiptBalanceOf1).Inc()
		w.logger.Warn("inspector: unable to check prev balance", "op", op.Hash(), "tx", receipt.TxHash.String(), "error", err)
		return
	}

	nextBalance, err := w.balanceOf(ctx, op.FeeToken, txBlockNum)
	if err != nil {
		// We can't check the balance, so we can't do anything
		w.metrics.inspectReceiptFailed.With(w.metrics.failedInspectReceiptBalanceOf2).Inc()
		w.logger.Warn("inspector: unable to check next balance", "op", op.Hash(), "tx", receipt.TxHash.String(), "error", err)
		return
	}

	effectiveGasPrice, err := w.fetchEffectiveGasPrice(ctx, tx, receipt)
	if err != nil {
		w.metrics.inspectReceiptFailed.With(w.metrics.failedInspectReceiptEffectiveGasPrice).Inc()
		w.logger.Warn("inspector: unable to check effective gas price", "op", op.Hash(), "tx", receipt.TxHash.String(), "error", err)
		return
	}

	balanceDiff := new(big.Int).Sub(nextBalance, prevBalance)
	nativeUsed := new(big.Int).Mul(effectiveGasPrice, big.NewInt(int64(receipt.GasUsed)))

	isNative := op.FeeToken == common.Address{}
	if isNative {
		balanceDiffFloat, _ := balanceDiff.Float64()
		balanceDiffFloat = math.Abs(balanceDiffFloat)

		if balanceDiff.Sign() == 1 {
			// We got paid, end of story
			w.metrics.overpaidAmount.Observe(balanceDiffFloat)

			w.logger.Info(
				"inspector: operation paid",
				"op", op.Hash(),
				"tx", receipt.TxHash.String(),
				"amount", balanceDiff.String(),
			)
			return
		}

		w.metrics.underpaidAmount.Observe(balanceDiffFloat)
		w.logger.Warn(
			"inspector: operation did not paid enough",
			"op", op.Hash(),
			"tx", receipt.TxHash.String(),
			"amount", balanceDiff.String(),
		)
	} else {
		// This is a bit more complicated, since we need to convert
		// the balanceDiff to native token and compare it with the nativeUsed
		nativePaid := priceSnap.ToNative(balanceDiff)
		nativeDiff := new(big.Int).Sub(nativePaid, nativeUsed)
		nativeDiffFloat, _ := nativeDiff.Float64()
		nativeDiffFloat = math.Abs(nativeDiffFloat)

		if nativePaid.Cmp(nativeUsed) >= 0 {
			// We got paid, end of story
			w.metrics.overpaidAmount.Observe(nativeDiffFloat)
			w.logger.Info(
				"inspector: operation paid",
				"op", op.Hash(),
				"tx", receipt.TxHash.String(),
				"token", op.FeeToken,
				"amount", balanceDiff.String(),
			)
			return
		}

		w.metrics.underpaidAmount.Observe(nativeDiffFloat)
		w.logger.Warn(
			"inspector: operation did not paid enough",
			"op", op.Hash(),
			"tx", receipt.TxHash.String(),
			"token", op.FeeToken,
			"amount", balanceDiff.String(),
			"nativePaid", nativePaid.String(),
			"nativeUsed", nativeUsed.String(),
		)
	}

	// The endorser lied to us
	w.ban <- &BanEndorser{Endorser: op.Endorser, Type: registry.PermanentBan}
}

func (w *Worker) balanceOf(ctx context.Context, token common.Address, blockNum *big.Int) (*big.Int, error) {
	isNative := token == common.Address{}
	if isNative {
		return w.Provider.BalanceAt(ctx, w.wallet.Address(), blockNum)
	}

	// Fetch from ERC20 balanceOf
	tokenContract, err := abierc20.NewERC20Caller(token, w.Provider)
	if err != nil {
		return nil, fmt.Errorf("unable to create ERC20Caller: %w", err)
	}

	return tokenContract.BalanceOf(&bind.CallOpts{
		BlockNumber: blockNum,
	}, w.wallet.Address())
}

func (w *Worker) fetchEffectiveGasPrice(
	ctx context.Context,
	tx *ethtypes.Transaction,
	receipt *ethtypes.Receipt,
) (*big.Int, error) {
	// If it exists on the receipt, we can use it
	if receipt.EffectiveGasPrice != nil {
		return receipt.EffectiveGasPrice, nil
	}

	// If not, we need to compute it.
	// it is the baseFee of the block + the priorityFee
	block, err := w.Provider.BlockByHash(ctx, receipt.BlockHash)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch block by hash: %w", err)
	}

	baseFee := new(big.Int).SetUint64(block.BaseFee().Uint64())
	priorityFee := new(big.Int).SetUint64(tx.GasTipCap().Uint64())

	return new(big.Int).Add(baseFee, priorityFee), nil
}

func (w *Worker) isOperationReady(
	ctx context.Context,
	op *types.Operation,
) (bool, error) {
	// We can check if the endorser is still returning `isValid == true`
	// since that would be a clear lie
	res, err := w.Endorser.IsOperationReady(ctx, op)
	if err != nil {
		return false, err
	}

	if !res.Readiness {
		return false, nil
	}

	ok, err := w.Endorser.ConstraintsMet(ctx, res)
	if err != nil {
		return false, err
	}

	return ok, nil
}
