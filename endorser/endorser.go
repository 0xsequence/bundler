package endorser

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/0xsequence/bundler/contracts/gen/solabis/abiendorser"
	"github.com/0xsequence/bundler/debugger"
	"github.com/0xsequence/bundler/types"
	"github.com/0xsequence/ethkit/ethcontract"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/common/hexutil"
	"github.com/go-chi/httplog/v2"
	"github.com/prometheus/client_golang/prometheus"
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

type Endorser struct {
	parsedEndorserABI *abi.ABI
	logger            *httplog.Logger
	metrics           *metrics

	Debugger debugger.Interface
	Provider *ethrpc.Provider
}

var _ Interface = (*Endorser)(nil)

func NewEndorser(logger *httplog.Logger, metrics prometheus.Registerer, provider *ethrpc.Provider, debugger debugger.Interface) *Endorser {
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
	rpcCall := ethrpc.NewCallBuilder[string]("eth_call", nil, endorserCall, nil, nil)
	_, err = e.Provider.Do(ctx, rpcCall.Into(&res))
	if err != nil {
		return nil, err
	}

	settingsResult, err := e.parseSimulationSettingsRes(res)
	if err != nil {
		return nil, fmt.Errorf("unable to parse result: %w", err)
	}

	return settingsResult, nil
}

func (e *Endorser) SimulationSettings(ctx context.Context, endorserAddr common.Address) ([]*SimulationSetting, error) {
	return e.simulationSettingsCall(ctx, endorserAddr)
}

func (e *Endorser) debugContextArgs(ctx context.Context, endorserAddr common.Address) (common.Address, *debugger.DebugContextArgs, error) {
	settings, err := e.simulationSettingsCall(ctx, endorserAddr)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("unable to get simulation settings: %w", err)
	}
	contextArgs := &debugger.DebugContextArgs{
		CodeReplacements: make([]debugger.CodeReplacement, 0, len(settings)),
		SlotReplacements: make([]debugger.SlotReplacement, 0, len(settings)),
	}
	to := endorserAddr
	for _, setting := range settings {
		if (setting.OldAddr != setting.NewAddr) {
			replacementCode, err := e.Provider.CodeAt(ctx, setting.NewAddr, nil)
			if err != nil {
				return common.Address{}, nil, fmt.Errorf("unable to read code for %v: %w", setting.OldAddr, err)
			}
			contextArgs.CodeReplacements = append(contextArgs.CodeReplacements, debugger.CodeReplacement{
				Address: setting.NewAddr,
				Code:    replacementCode,
			})
			if (setting.OldAddr != endorserAddr) {
				// Update the endorser location
				to = setting.NewAddr
			}
		}
		for _, slot := range setting.Slots {
			contextArgs.SlotReplacements = append(contextArgs.SlotReplacements, debugger.SlotReplacement{
				Address: setting.NewAddr,
				Slot:    slot.Slot,
				Value:   slot.Value,
			})
		}
	}
	return to, contextArgs, nil
}

// IsOperationReady

func (e *Endorser) buildIsOperationReadyCalldata(op *types.Operation) (common.Address, string, error) {
	endorser := ethcontract.NewContractCaller(op.Endorser, *e.parsedEndorserABI, e.Provider)
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

	to, data, err := e.buildIsOperationReadyCalldata(op)
	if err != nil {
		e.metrics.isOperationReadyError.Inc()
		return nil, fmt.Errorf("unable to build calldata: %w", err)
	}

	endorserCall := &struct {
		To   common.Address `json:"to"`
		Data string         `json:"data"`
	}{
		To:   to,
		Data: data,
	}

	var res string
	rpcCall := ethrpc.NewCallBuilder[string]("eth_call", nil, endorserCall)
	_, err = e.Provider.Do(ctx, rpcCall.Into(&res))
	if err != nil {
		e.metrics.isOperationReadyError.Inc()
		return nil, err
	}

	endorserResult, err := e.parseIsOperationReadyRes(res)
	if err != nil {
		e.metrics.isOperationReadyReverts.Inc()
		return nil, fmt.Errorf("unable to parse result: %w", err)
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

	to, data, err := e.buildIsOperationReadyCalldata(op)
	if err != nil {
		return nil, fmt.Errorf("unable to build calldata: %w", err)
	}

	to, debugContextArgs, err := e.debugContextArgs(ctx, to)
	if err != nil {
		return nil, fmt.Errorf("unable to build debug context args: %w", err)
	}

	// Use random caller
	// NOTICE: This is a temporary solution
	debugCallArgs := &debugger.DebugCallArgs{
		From: common.HexToAddress("0xFD095316B59e6224dC84f83E68F9603A684AD8df"),
		To:   to,
		Data: common.FromHex(data),
	}

	trace, err := e.Debugger.DebugTraceCallContext(ctx, debugCallArgs, debugContextArgs)
	if err != nil {
		return nil, fmt.Errorf("unable to trace call: %w", err)
	}

	er1, err := e.parseIsOperationReadyRes(trace.ReturnValue)
	if err != nil {
		return nil, fmt.Errorf("unable to parse result: %w", err)
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
		e.logger.Warn("unable to use debugger, falling back to eth_call and ignoring simulation settings", "error", err)
	}

	return e.isOperationReadyCall(ctx, op)
}

func (e *Endorser) DependencyState(ctx context.Context, result *EndorserResult) (*EndorserResultState, error) {
	start := time.Now()
	state := EndorserResultState{}

	state.AddrDependencies = make(map[common.Address]*AddrDependencyState, len(result.Dependencies))

	for _, dependency := range result.Dependencies {
		state_ := AddrDependencyState{}

		if dependency.Balance {
			var err error
			state_.Balance, err = e.Provider.BalanceAt(ctx, dependency.Addr, nil)
			if err != nil {
				e.metrics.dependencyStateError.With(e.metrics.dependencyStateErrorBalance).Inc()
				return nil, fmt.Errorf("unable to read balance for %v: %w", dependency.Addr, err)
			}
		}

		if dependency.Code {
			code, err := e.Provider.CodeAt(ctx, dependency.Addr, nil)
			if err != nil {
				e.metrics.dependencyStateError.With(e.metrics.dependencyStateErrorCode).Inc()
				return nil, fmt.Errorf("unable to read code for %v: %w", dependency.Addr, err)
			}
			if code == nil {
				code = []byte{}
			}
			state_.Code = code
		}

		if dependency.Nonce {
			nonce, err := e.Provider.NonceAt(ctx, dependency.Addr, nil)
			if err != nil {
				e.metrics.dependencyStateError.With(e.metrics.dependencyStateErrorNonce).Inc()
				return nil, fmt.Errorf("unable to read nonce for %v: %w", dependency.Addr, err)
			}
			state_.Nonce = &nonce
		}

		state_.Slots = make([][32]byte, 0, len(dependency.Slots))
		for _, slot := range dependency.Slots {
			start2 := time.Now()
			value, err := e.Provider.StorageAt(ctx, dependency.Addr, slot, nil)
			e.metrics.dependencySlotDuration.Observe(time.Since(start2).Seconds())

			if err != nil {
				e.metrics.dependencyStateError.With(e.metrics.dependencyStateErrorSlots).Inc()
				return nil, fmt.Errorf("unable to read storage for %v at %v: %w", dependency.Addr, hexutil.Encode(slot[:]), err)
			}

			state_.Slots = append(state_.Slots, [32]byte(value))
		}

		state.AddrDependencies[dependency.Addr] = &state_
	}

	e.metrics.dependencyStateDuration.Observe(time.Since(start).Seconds())
	return &state, nil
}

func (e *Endorser) ConstraintsMet(ctx context.Context, result *EndorserResult) (bool, error) {
	start := time.Now()

	for _, dependency := range result.Dependencies {
		for _, constraint := range dependency.Constraints {
			start2 := time.Now()
			ok, err := CheckConstraint(ctx, e.Provider, dependency.Addr, constraint.Slot, constraint.MinValue, constraint.MaxValue)
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

func CheckConstraint(ctx context.Context, provider *ethrpc.Provider, addr common.Address, slot [32]byte, minValue, maxValue [32]byte) (bool, error) {
	value, err := provider.StorageAt(ctx, addr, slot, nil)
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
