package endorser

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/0xsequence/bundler/contracts/gen/solabis/abiendorser"
	"github.com/0xsequence/bundler/lib/debugger"
	"github.com/0xsequence/bundler/lib/provider"
	"github.com/0xsequence/bundler/lib/types"
	"github.com/0xsequence/ethkit/ethcontract"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/common/hexutil"
	"github.com/go-chi/httplog/v2"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
)

var parsedEndorserABI *abi.ABI

func useEndorserAbi() *abi.ABI {
	if parsedEndorserABI != nil {
		return parsedEndorserABI
	}

	parsed := ethcontract.MustParseABI(abiendorser.EndorserABI)
	parsedEndorserABI = &parsed
	return parsedEndorserABI
}

type rpcError struct {
	Err error
}

func (re *rpcError) Error() string {
	return re.Err.Error()
}

type rejectedError struct {
	Err    error
	Reason string
}

func (re *rejectedError) Error() string {
	return fmt.Sprintf("rejected: %s", re.Reason)
}

type Endorser struct {
	parsedEndorserABI *abi.ABI
	logger            *httplog.Logger
	metrics           *metrics

	Debugger debugger.Interface
	Provider *provider.Batched
}

var _ Interface = (*Endorser)(nil)

func NewEndorser(logger *httplog.Logger, metrics prometheus.Registerer, provider *provider.Batched, debugger debugger.Interface) *Endorser {
	return &Endorser{
		parsedEndorserABI: useEndorserAbi(),

		logger:   logger,
		metrics:  createMetrics(metrics),
		Debugger: debugger,
		Provider: provider,
	}
}

// SimulationSettings

func (e *Endorser) parseSimulationSettingsRes(res string) ([]*SimulationSetting, error) {
	resBytes := common.FromHex(res)

	settingsResult := []*SimulationSetting{}

	values, err := e.parsedEndorserABI.Methods["simulationSettings"].Outputs.Unpack(resBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to unpack result: %w", err)
	}

	// Must be an array of structs
	vals, ok := values[0].([]struct {
		OldAddr common.Address "json:\"oldAddr\""
		NewAddr common.Address "json:\"newAddr\""
		Slots   []struct {
			Slot  [32]byte "json:\"slot\""
			Value [32]byte "json:\"value\""
		} "json:\"slots\""
	})
	if !ok {
		return nil, fmt.Errorf("invalid settings")
	}

	for _, val := range vals {
		setting := &SimulationSetting{
			OldAddr: val.OldAddr,
			NewAddr: val.NewAddr,
			Slots:   make([]SlotReplacement, 0, len(val.Slots)),
		}
		for _, slot := range val.Slots {
			setting.Slots = append(setting.Slots, SlotReplacement{
				Slot:  slot.Slot,
				Value: slot.Value,
			})
		}
		settingsResult = append(settingsResult, setting)
	}

	return settingsResult, nil
}

func (e *Endorser) simulationSettingsCall(ctx context.Context, endorserAddr common.Address) ([]*SimulationSetting, error) {
	endorser := ethcontract.NewContractCaller(endorserAddr, *e.parsedEndorserABI, e.Provider)
	calldata, err := endorser.Encode("simulationSettings")

	if err != nil {
		return nil, err
	}

	endorserCall := &struct {
		To   common.Address `json:"to"`
		Data string         `json:"data"`
	}{
		To:   endorserAddr,
		Data: "0x" + common.Bytes2Hex(calldata),
	}

	var res string
	rpcCall := ethrpc.NewCallBuilder[string]("eth_call", nil, endorserCall)
	_, err = e.Provider.Do(ctx, rpcCall.Into(&res))
	if err != nil {
		return nil, err
	}

	e.logger.Debug("simulation settings call", "res", res)

	settingsResult, err := e.parseSimulationSettingsRes(res)
	if err != nil {
		return nil, fmt.Errorf("unable to parse simulation settings result: %w", err)
	}

	return settingsResult, nil
}

func (e *Endorser) SimulationSettings(ctx context.Context, endorserAddr common.Address) ([]*SimulationSetting, error) {
	return e.simulationSettingsCall(ctx, endorserAddr)
}

func (e *Endorser) callOverrideArgs(ctx context.Context, endorserAddr common.Address) (common.Address, *debugger.DebugOverrideArgs, error) {
	settings, err := e.simulationSettingsCall(ctx, endorserAddr)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("unable to get simulation settings: %w", err)
	}
	overrideArgs := debugger.DebugOverrideArgs{}
	to := endorserAddr
	for _, setting := range settings {
		if setting.OldAddr == endorserAddr {
			// Update the endorser location
			to = setting.NewAddr
		}

		overrideArg := debugger.DebugOverride{}
		if setting.OldAddr != setting.NewAddr {
			// Code replacement
			replacementCode, err := e.Provider.CodeAt(ctx, setting.OldAddr, nil)
			if err != nil {
				return common.Address{}, nil, fmt.Errorf("unable to read code for %v: %w", setting.OldAddr, err)
			}
			codeStr := "0x" + common.Bytes2Hex(replacementCode)
			overrideArg.Code = &codeStr
		}
		// Slots
		if len(setting.Slots) > 0 {
			overrideArg.StateDiff = make(map[common.Hash]common.Hash)
		}
		for _, slot := range setting.Slots {
			overrideArg.StateDiff[common.BytesToHash(slot.Slot[:])] = common.BytesToHash(slot.Value[:])
		}

		overrideArgs[setting.NewAddr] = &overrideArg
	}
	return to, &overrideArgs, nil
}

// IsOperationReady

func (e *Endorser) BuildIsOperationReadyCalldata(op *types.Operation) (common.Address, string, error) {
	endorser := ethcontract.NewContractCaller(op.Endorser, *e.parsedEndorserABI, nil)
	calldata, err := endorser.Encode("isOperationReady", &op.IEndorserOperation)

	if err != nil {
		return common.Address{}, "", err
	}

	return op.Endorser, "0x" + common.Bytes2Hex(calldata), nil
}

func (e *Endorser) parseIsOperationReadyRes(res string) (*EndorserResult, error) {
	resBytes := common.FromHex(res)

	endorserResult := &EndorserResult{}

	dec1, err := e.parsedEndorserABI.Methods["isOperationReady"].Outputs.Unpack(resBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to unpack result: %w", err)
	}

	// It must have 3 elements
	if len(dec1) != 3 {
		return nil, fmt.Errorf("invalid result length")
	}

	// First element must be a bool
	ready, ok := dec1[0].(bool)
	if !ok {
		return nil, fmt.Errorf("invalid readiness")
	}

	endorserResult.Readiness = ready

	// Second element must be a struct
	dec2, ok := dec1[1].(struct {
		BaseFee           bool     `json:"baseFee"`
		BlobBaseFee       bool     `json:"blobBaseFee"`
		ChainId           bool     `json:"chainId"`
		CoinBase          bool     `json:"coinBase"`
		Difficulty        bool     `json:"difficulty"`
		GasLimit          bool     `json:"gasLimit"`
		Number            bool     `json:"number"`
		Timestamp         bool     `json:"timestamp"`
		TxOrigin          bool     `json:"txOrigin"`
		TxGasPrice        bool     `json:"txGasPrice"`
		MaxBlockNumber    *big.Int `json:"maxBlockNumber"`
		MaxBlockTimestamp *big.Int `json:"maxBlockTimestamp"`
	})
	if !ok {
		return nil, fmt.Errorf("invalid block dependency")
	}

	endorserResult.GlobalDependency = abiendorser.IEndorserGlobalDependency{
		BaseFee:           dec2.BaseFee,
		BlobBaseFee:       dec2.BlobBaseFee,
		ChainId:           dec2.ChainId,
		CoinBase:          dec2.CoinBase,
		Difficulty:        dec2.Difficulty,
		GasLimit:          dec2.GasLimit,
		Number:            dec2.Number,
		Timestamp:         dec2.Timestamp,
		TxOrigin:          dec2.TxOrigin,
		TxGasPrice:        dec2.TxGasPrice,
		MaxBlockNumber:    dec2.MaxBlockNumber,
		MaxBlockTimestamp: dec2.MaxBlockTimestamp,
	}

	// Third element must be an array of structs
	dec3, ok := dec1[2].([]struct {
		Addr        common.Address "json:\"addr\""
		Balance     bool           "json:\"balance\""
		Code        bool           "json:\"code\""
		Nonce       bool           "json:\"nonce\""
		AllSlots    bool           "json:\"allSlots\""
		Slots       [][32]byte     "json:\"slots\""
		Constraints []struct {
			Slot     [32]byte "json:\"slot\""
			MinValue [32]byte "json:\"minValue\""
			MaxValue [32]byte "json:\"maxValue\""
		} "json:\"constraints\""
	})
	if !ok {
		return nil, fmt.Errorf("invalid dependencies")
	}

	endorserResult.Dependencies = make([]abiendorser.IEndorserDependency, 0, len(dec3))
	for _, dep := range dec3 {
		dependency := abiendorser.IEndorserDependency{
			Addr:     dep.Addr,
			Balance:  dep.Balance,
			Code:     dep.Code,
			Nonce:    dep.Nonce,
			AllSlots: dep.AllSlots,
			Slots:    dep.Slots,
		}
		dependency.Constraints = make([]abiendorser.IEndorserConstraint, 0, len(dep.Constraints))
		for _, c := range dep.Constraints {
			dependency.Constraints = append(dependency.Constraints, abiendorser.IEndorserConstraint{
				Slot:     c.Slot,
				MinValue: c.MinValue,
				MaxValue: c.MaxValue,
			})
		}
		endorserResult.Dependencies = append(endorserResult.Dependencies, dependency)
	}

	return endorserResult, nil
}

func (e *Endorser) isOperationReadyCall(ctx context.Context, op *types.Operation) (*EndorserResult, error) {
	start := time.Now()
	e.metrics.isOperationReadyAttempts.Inc()

	to, data, err := e.BuildIsOperationReadyCalldata(op)
	if err != nil {
		e.metrics.isOperationReadyError.Inc()
		return nil, fmt.Errorf("unable to build calldata: %w", err)
	}

	to, debugOverrideArgs, err := e.callOverrideArgs(ctx, to)
	if err != nil {
		return nil, fmt.Errorf("unable to build debug override args: %w", err)
	}

	endorserCall := &struct {
		To   common.Address `json:"to"`
		Data string         `json:"data"`
	}{
		To:   to,
		Data: data,
	}

	var res string
	// Some RPC providers may not support the debugOverrideArgs
	rpcCall := ethrpc.NewCallBuilder[string]("eth_call", nil, endorserCall, "latest", debugOverrideArgs)
	_, err = e.Provider.Do(ctx, rpcCall.Into(&res))
	if err != nil {
		e.metrics.isOperationReadyError.Inc()
		if strings.Contains(err.Error(), "execution reverted") {
			reason := err.Error()
			reason = strings.TrimPrefix(reason, "jsonrpc error 3: ")
			reason = strings.TrimPrefix(reason, "execution reverted: ")
			reason = strings.TrimPrefix(reason, "reverted: ")
			return nil, &rejectedError{Err: err, Reason: reason}
		}

		return nil, &rpcError{Err: err}
	}

	endorserResult, err := e.parseIsOperationReadyRes(res)
	if err != nil {
		e.metrics.isOperationReadyReverts.Inc()
		return nil, fmt.Errorf("unable to parse isOperationReady result: %w", err)
	}

	// NOTICE: Untrusted context operations should be handled
	// by the debugger, but if it's not available we still handle
	// them, we just mark them as wildcard only.
	if op.HasUntrustedContext {
		e.metrics.isOperationReadyWildcards.Inc()
		endorserResult.WildcardOnly = true
	}

	if endorserResult.Readiness {
		e.metrics.isOperationReadyTrue.Inc()
	} else {
		e.metrics.isOperationReadyFalse.Inc()
	}

	e.metrics.isOperationReadyDuration.Observe(time.Since(start).Seconds())
	egl, _ := op.EndorserGasLimit.Float64()
	e.metrics.durationPerGas.Observe(time.Since(start).Seconds() / egl)

	return endorserResult, nil
}

func (e *Endorser) isOperationReadyDebugger(ctx context.Context, op *types.Operation) (*EndorserResult, error) {
	start := time.Now()

	e.metrics.isOperationReadyDebugger.Inc()
	if e.Debugger == nil {
		return nil, fmt.Errorf("debugger is not available")
	}

	to, data, err := e.BuildIsOperationReadyCalldata(op)
	if err != nil {
		return nil, fmt.Errorf("unable to build calldata: %w", err)
	}

	to, debugOverrideArgs, err := e.callOverrideArgs(ctx, to)
	if err != nil {
		return nil, fmt.Errorf("unable to build debug override args: %w", err)
	}

	// Use random caller
	// NOTICE: This is a temporary solution
	debugCallArgs := &debugger.DebugCallArgs{
		From: common.HexToAddress("0xFD095316B59e6224dC84f83E68F9603A684AD8df"),
		To:   to,
		Data: common.FromHex(data),
	}

	trace, err := e.Debugger.DebugTraceCall(ctx, debugCallArgs, debugOverrideArgs)
	if err != nil {
		return nil, fmt.Errorf("unable to trace call: %w", err)
	}
	// Log the trace
	for _, log := range trace.StructLogs {
		// If Op starts with "LOG", then it's a debug trace log
		if len(log.Op) > 3 && log.Op[:3] == "LOG" {
			e.logger.Debug("debug trace log", "log", log)
		}
	}

	er1, err := e.parseIsOperationReadyRes(trace.ReturnValue)
	if err != nil {
		return nil, fmt.Errorf("unable to parse isOperationReady debugger result: %w", err)
	}

	// Generate dependencies from untrusted context
	er2, err := ParseUntrustedDebug(trace)
	if err != nil {
		return nil, fmt.Errorf("unable to parse untrusted debug: %w", err)
	}

	merged := er1.Or(er2)

	e.metrics.isOperationDebugReadyDuration.Observe(time.Since(start).Seconds())
	egl, _ := op.EndorserGasLimit.Float64()
	e.metrics.debugDurationPerGas.Observe(time.Since(start).Seconds() / egl)

	return merged, nil
}

func (e *Endorser) simulateCall(ctx context.Context, op *types.Operation) (*EndorserResult, error) {
	start := time.Now()
	e.metrics.failoverSimulationAttempts.Inc()

	txData := &struct {
		To   common.Address `json:"to"`
		Data string         `json:"data"`
	}{
		To:   op.Entrypoint,
		Data: "0x" + common.Bytes2Hex(op.Data),
	}

	e.logger.Debug("simulation call", "to", txData.To, "data", txData.Data)

	var res string
	rpcCall := ethrpc.NewCallBuilder[string]("eth_call", nil, txData)
	_, err := e.Provider.Do(ctx, rpcCall.Into(&res))
	if err != nil {
		e.metrics.failoverSimulationError.Inc()
		return nil, fmt.Errorf("unable to simulate call: %w", err)
	}
	e.metrics.failoverSimulationSuccess.Inc()

	e.logger.Debug("simulation call success", "res", res)

	e.metrics.failoverSimulationDuration.Observe(time.Since(start).Seconds())
	gasLimit := big.NewInt(0).Add(op.FixedGas, op.GasLimit)
	limitFloat, _ := gasLimit.Float64()
	e.metrics.durationPerGas.Observe(time.Since(start).Seconds() / limitFloat)

	// We can't get the dependencies from the simulation, so we just mark it as wildcard only
	e.metrics.isOperationReadyWildcards.Inc()
	return &EndorserResult{
		Readiness:    true,
		WildcardOnly: true,
	}, nil
}

func (e *Endorser) IsOperationReady(ctx context.Context, op *types.Operation) (*EndorserResult, error) {
	if e.Debugger != nil && op.HasUntrustedContext {
		// TODO: Sometimes the endorser reverts instead of failing,
		// we should have a different sort of error for these, or else
		// we will have to verify them twice.
		res, err := e.isOperationReadyDebugger(ctx, op)
		if err == nil {
			return res, nil
		}

		e.metrics.isOperationReadyDebuggerFailed.Inc()
		e.logger.Warn("unable to use debugger, falling back to eth_call and ignoring untrusted context", "error", err)
	}

	// Use eth_call
	res, err := e.isOperationReadyCall(ctx, op)
	if err != nil {
		_, ok := err.(*rpcError)
		if ok {
			// Fail over to pure calldata simulation
			e.logger.Warn("unable to use endorser, falling back to call simulation and ignoring dependencies", "error", err)
			return e.simulateCall(ctx, op)
		}

		return nil, err
	}

	return res, nil
}

func (e *Endorser) SingleDependencyState(ctx context.Context, dep abiendorser.IEndorserDependency) (*AddrDependencyState, error) {
	balance := make(chan *big.Int, 1)
	code := make(chan int, 1)
	nonce := make(chan uint64, 1)
	slots := make(chan [][32]byte, 1)

	defer func() {
		close(balance)
		close(nonce)
		close(code)
		close(slots)
	}()

	eg, ctx := errgroup.WithContext(ctx)

	if dep.Balance {
		eg.Go(func() error {
			b, err := e.Provider.BalanceAt(ctx, dep.Addr, nil)
			if err != nil {
				return fmt.Errorf("unable to read balance for %v: %w", dep.Addr, err)
			}

			balance <- b
			return nil
		})
	}

	if dep.Code {
		eg.Go(func() error {
			c, err := e.Provider.CodeAt(ctx, dep.Addr, nil)
			if err != nil {
				return fmt.Errorf("unable to read code for %v: %w", dep.Addr, err)
			}

			code <- len(c)
			return nil
		})
	}

	if dep.Nonce {
		eg.Go(func() error {
			n, err := e.Provider.NonceAt(ctx, dep.Addr, nil)
			if err != nil {
				return fmt.Errorf("unable to read nonce for %v: %w", dep.Addr, err)
			}

			nonce <- n
			return nil
		})
	}

	if len(dep.Slots) > 0 {
		eg.Go(func() error {
			s, err := e.Provider.StorageAtBatch(ctx, dep.Addr, dep.Slots)
			if err != nil {
				return fmt.Errorf("unable to read storage for %v: %w", dep.Addr, err)
			}

			slots <- s
			return nil
		})
	}

	err := eg.Wait()
	if err != nil {
		return nil, err
	}

	state := &AddrDependencyState{}

	for {
		select {
		case b := <-balance:
			state.Balance = b
		case n := <-nonce:
			state.Nonce = &n
		case c := <-code:
			state.Code = &c
		case s := <-slots:
			state.Slots = s
		default:
			return state, nil
		}
	}
}

func (e *Endorser) DependencyState(ctx context.Context, result *EndorserResult) (*EndorserResultState, error) {
	type SingleDependencyStateResult struct {
		Addr common.Address
		Res  *AddrDependencyState
	}

	deps := make(chan *SingleDependencyStateResult, len(result.Dependencies))
	defer close(deps)

	eg, ctx := errgroup.WithContext(ctx)

	for _, dep := range result.Dependencies {
		capturedDep := dep
		eg.Go(func() error {
			res, err := e.SingleDependencyState(ctx, capturedDep)
			if err != nil {
				return err
			}

			deps <- &SingleDependencyStateResult{Addr: capturedDep.Addr, Res: res}
			return nil
		})
	}

	err := eg.Wait()
	if err != nil {
		return nil, err
	}

	res := make(map[common.Address]*AddrDependencyState, len(result.Dependencies))
	for {
		select {
		case dep := <-deps:
			res[dep.Addr] = dep.Res
		default:
			return &EndorserResultState{AddrDependencies: res}, nil
		}
	}
}

func (e *Endorser) ConstraintsMet(ctx context.Context, result *EndorserResult) (bool, error) {
	start := time.Now()

	for _, dependency := range result.Dependencies {
		for _, constraint := range dependency.Constraints {
			start2 := time.Now()
			ok, err := e.CheckConstraint(ctx, dependency.Addr, constraint.Slot, constraint.MinValue, constraint.MaxValue)
			e.metrics.constraintMetDuration.Observe(time.Since(start2).Seconds())

			if err != nil {
				e.metrics.constraintsMetError.Inc()
				return false, err
			}

			if !ok {
				e.metrics.constraintsNotMet.Inc()
				return false, nil
			}
		}
	}

	e.metrics.constraintsMet.Inc()
	e.metrics.constraintsMetDuration.Observe(time.Since(start).Seconds())
	return true, nil
}

func (e *Endorser) CheckConstraint(ctx context.Context, addr common.Address, slot [32]byte, minValue, maxValue [32]byte) (bool, error) {
	value, err := e.Provider.StorageAt(ctx, addr, slot, nil)
	if err != nil {
		return false, fmt.Errorf("unable to read storage for %v at %v: %w", addr, hexutil.Encode(slot[:]), err)
	}

	bnMin := new(big.Int).SetBytes(minValue[:])
	bnMax := new(big.Int).SetBytes(maxValue[:])
	bnValue := new(big.Int).SetBytes(value[:])

	if bnValue.Cmp(bnMin) < 0 || bnValue.Cmp(bnMax) > 0 {
		return false, nil
	}

	return true, nil
}
