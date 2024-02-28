package endorser

import (
	"context"
	"fmt"
	"math/big"

	"github.com/0xsequence/bundler/contracts/gen/solabis/abiendorser"
	"github.com/0xsequence/bundler/types"
	"github.com/0xsequence/ethkit/ethcontract"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/common/hexutil"
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

	Provider *ethrpc.Provider
}

var _ Interface = (*Endorser)(nil)

func NewEndorser(provider *ethrpc.Provider) *Endorser {
	return &Endorser{
		parsedEndorserABI: useEndorserAbi(),
		Provider:          provider,
	}
}

func (e *Endorser) IsOperationReady(ctx context.Context, op *types.Operation) (*EndorserResult, error) {
	endorser := ethcontract.NewContractCaller(op.Endorser, *e.parsedEndorserABI, e.Provider)

	calldata, err := endorser.Encode(
		"isOperationReady",
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
	)

	if err != nil {
		return nil, err
	}

	type Call struct {
		To   common.Address `json:"to"`
		Data string         `json:"data"`
	}

	endorserCall := &Call{
		To:   op.Endorser,
		Data: "0x" + common.Bytes2Hex(calldata),
	}

	var res string
	rpcCall := ethrpc.NewCallBuilder[string]("eth_call", nil, endorserCall, nil, nil)
	_, err = e.Provider.Do(ctx, rpcCall.Into(&res))
	if err != nil {
		return nil, err
	}

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
		Basefee           bool     "json:\"basefee\""
		Blobbasefee       bool     "json:\"blobbasefee\""
		Chainid           bool     "json:\"chainid\""
		Coinbase          bool     "json:\"coinbase\""
		Difficulty        bool     "json:\"difficulty\""
		GasLimit          bool     "json:\"gasLimit\""
		Number            bool     "json:\"number\""
		Timestamp         bool     "json:\"timestamp\""
		TxOrigin          bool     "json:\"txOrigin\""
		TxGasPrice        bool     "json:\"txGasPrice\""
		MaxBlockNumber    *big.Int "json:\"maxBlockNumber\""
		MaxBlockTimestamp *big.Int "json:\"maxBlockTimestamp\""
	})
	if !ok {
		return nil, fmt.Errorf("invalid block dependency")
	}

	endorserResult.GlobalDependency = abiendorser.EndorserGlobalDependency{
		Basefee:           dec2.Basefee,
		Blobbasefee:       dec2.Blobbasefee,
		Chainid:           dec2.Chainid,
		Coinbase:          dec2.Coinbase,
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

	endorserResult.Dependencies = make([]abiendorser.EndorserDependency, 0, len(dec3))
	for _, dep := range dec3 {
		dependency := abiendorser.EndorserDependency{
			Addr:     dep.Addr,
			Balance:  dep.Balance,
			Code:     dep.Code,
			Nonce:    dep.Nonce,
			AllSlots: dep.AllSlots,
			Slots:    dep.Slots,
		}
		dependency.Constraints = make([]abiendorser.EndorserConstraint, 0, len(dep.Constraints))
		for _, c := range dep.Constraints {
			dependency.Constraints = append(dependency.Constraints, abiendorser.EndorserConstraint{
				Slot:     c.Slot,
				MinValue: c.MinValue,
				MaxValue: c.MaxValue,
			})
		}
		endorserResult.Dependencies = append(endorserResult.Dependencies, dependency)
	}

	return endorserResult, nil
}

func (e *Endorser) DependencyState(ctx context.Context, result *EndorserResult) (*EndorserResultState, error) {
	state := EndorserResultState{}

	state.AddrDependencies = make(map[common.Address]*AddrDependencyState, len(result.Dependencies))

	for _, dependency := range result.Dependencies {
		state_ := AddrDependencyState{}

		if dependency.Balance {
			var err error
			state_.Balance, err = e.Provider.BalanceAt(ctx, dependency.Addr, nil)
			if err != nil {
				return nil, fmt.Errorf("unable to read balance for %v: %w", dependency.Addr, err)
			}
		}

		if dependency.Code {
			code, err := e.Provider.CodeAt(ctx, dependency.Addr, nil)
			if err != nil {
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
				return nil, fmt.Errorf("unable to read nonce for %v: %w", dependency.Addr, err)
			}
			state_.Nonce = &nonce
		}

		state_.Slots = make([][32]byte, 0, len(dependency.Slots))
		for _, slot := range dependency.Slots {
			value, err := e.Provider.StorageAt(ctx, dependency.Addr, slot, nil)
			if err != nil {
				return nil, fmt.Errorf("unable to read storage for %v at %v: %w", dependency.Addr, hexutil.Encode(slot[:]), err)
			}
			state_.Slots = append(state_.Slots, [32]byte(value))
		}

		state.AddrDependencies[dependency.Addr] = &state_
	}

	return &state, nil
}

func (e *Endorser) ConstraintsMet(ctx context.Context, result *EndorserResult) (bool, error) {
	for _, dependency := range result.Dependencies {
		for _, constraint := range dependency.Constraints {
			ok, err := CheckConstraint(ctx, e.Provider, dependency.Addr, constraint.Slot, constraint.MinValue, constraint.MaxValue)
			if err != nil {
				return false, err
			}

			if !ok {
				return false, nil
			}
		}
	}

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
