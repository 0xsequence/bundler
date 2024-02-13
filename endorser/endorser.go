package endorser

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/ethkit/ethcontract"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/common/hexutil"
)

const ENDORSER_ABI = `
[
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "_entrypoint",
				"type": "address"
			},
			{
				"internalType": "bytes",
				"name": "_data",
				"type": "bytes"
			},
			{
				"internalType": "bytes",
				"name": "_endorserCallData",
				"type": "bytes"
			},
			{
				"internalType": "uint256",
				"name": "_gasLimit",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "_maxFeePerGas",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "_maxPriorityFeePerGas",
				"type": "uint256"
			},
			{
				"internalType": "address",
				"name": "_feeToken",
				"type": "address"
			}
		],
		"name": "isOperationReady",
		"outputs": [
			{
				"internalType": "bool",
				"name": "readiness",
				"type": "bool"
			},
			{
				"components": [
					{
						"internalType": "uint256",
						"name": "maxNumber",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "maxTimestamp",
						"type": "uint256"
					}
				],
				"internalType": "struct Endorser.BlockDependency",
				"name": "blockDependency",
				"type": "tuple"
			},
			{
				"components": [
					{
						"internalType": "address",
						"name": "addr",
						"type": "address"
					},
					{
						"internalType": "bool",
						"name": "balance",
						"type": "bool"
					},
					{
						"internalType": "bool",
						"name": "code",
						"type": "bool"
					},
					{
						"internalType": "bool",
						"name": "nonce",
						"type": "bool"
					},
					{
						"internalType": "bool",
						"name": "allSlots",
						"type": "bool"
					},
					{
						"internalType": "bytes32[]",
						"name": "slots",
						"type": "bytes32[]"
					},
					{
						"components": [
							{
								"internalType": "bytes32",
								"name": "slot",
								"type": "bytes32"
							},
							{
								"internalType": "bytes32",
								"name": "minValue",
								"type": "bytes32"
							},
							{
								"internalType": "bytes32",
								"name": "maxValue",
								"type": "bytes32"
							}
						],
						"internalType": "struct Endorser.Constraint[]",
						"name": "constraints",
						"type": "tuple[]"
					}
				],
				"internalType": "struct Endorser.Dependency[]",
				"name": "dependencies",
				"type": "tuple[]"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]
`

type BlockDependency struct {
	MaxNumber    *big.Int
	MaxTimestamp *big.Int
}

type Constraint struct {
	Slot     [32]byte
	MinValue [32]byte
	MaxValue [32]byte
}

type Dependency struct {
	Addr       common.Address
	Balance    bool
	Code       bool
	Nonce      bool
	AllSlots   bool
	Slots      [][32]byte
	Constraint []Constraint
}

func (d *Dependency) HasChanged(x, y *dependencyState) (bool, error) {
	if err := d.Validate(x); err != nil {
		return false, fmt.Errorf("x is not a valid state for dependency on %v: %w", d.Addr, err)
	}
	if err := d.Validate(y); err != nil {
		return false, fmt.Errorf("y is not a valid state for dependency on %v: %w", d.Addr, err)
	}

	if d.Balance {
		if x.Balance.Cmp(y.Balance) != 0 {
			return true, nil
		}
	}

	if d.Code {
		if !bytes.Equal(x.Code, y.Code) {
			return true, nil
		}
	}

	if d.Nonce {
		if *x.Nonce != *y.Nonce {
			return true, nil
		}
	}

	return false, nil
}

func (d *Dependency) Validate(state *dependencyState) error {
	if (state.Balance != nil) != d.Balance {
		return fmt.Errorf("balance existence does not match dependency")
	}

	if (state.Code != nil) != d.Code {
		return fmt.Errorf("code existence does not match dependency")
	}

	if (state.Nonce != nil) != d.Nonce {
		return fmt.Errorf("nonce existence does not match dependency")
	}

	if len(state.Slots) != len(d.Slots) {
		return fmt.Errorf("number of slots does not match dependency")
	}

	return nil
}

type dependencyState struct {
	Balance *big.Int
	Code    []byte
	Nonce   *uint64
	Slots   [][32]byte
}

type EndorserResult struct {
	Readiness       bool
	BlockDependency BlockDependency
	Dependencies    []Dependency
}

func (r *EndorserResult) State(ctx context.Context, provider *ethrpc.Provider) (*endorserResultState, error) {
	state := endorserResultState{}

	for _, dependency := range r.Dependencies {
		state_ := dependencyState{}

		if dependency.Balance {
			var err error
			state_.Balance, err = provider.BalanceAt(ctx, dependency.Addr, nil)
			if err != nil {
				return nil, fmt.Errorf("unable to read balance for %v: %w", dependency.Addr, err)
			}
		}

		if dependency.Code {
			code, err := provider.CodeAt(ctx, dependency.Addr, nil)
			if err != nil {
				return nil, fmt.Errorf("unable to read code for %v: %w", dependency.Addr, err)
			}
			if code == nil {
				code = []byte{}
			}
			state_.Code = code
		}

		if dependency.Nonce {
			nonce, err := provider.NonceAt(ctx, dependency.Addr, nil)
			if err != nil {
				return nil, fmt.Errorf("unable to read nonce for %v: %w", dependency.Addr, err)
			}
			state_.Nonce = &nonce
		}

		state_.Slots = make([][32]byte, 0, len(dependency.Slots))
		for _, slot := range dependency.Slots {
			value, err := provider.StorageAt(ctx, dependency.Addr, slot, nil)
			if err != nil {
				return nil, fmt.Errorf("unable to read storage for %v at %v: %w", dependency.Addr, hexutil.Encode(slot[:]), err)
			}
			state_.Slots = append(state_.Slots, [32]byte(value))
		}

		state.Dependencies = append(state.Dependencies, state_)
	}

	return &state, nil
}

func (r *EndorserResult) HasChanged(x, y *endorserResultState) (bool, error) {
	if err := r.Validate(x); err != nil {
		return false, fmt.Errorf("x is not a valid state for endorser result: %w", err)
	}
	if err := r.Validate(y); err != nil {
		return false, fmt.Errorf("y is not a valid state for endorser result: %w", err)
	}

	for i, dependency := range r.Dependencies {
		hasChanged, err := dependency.HasChanged(&x.Dependencies[i], &y.Dependencies[i])
		if err != nil {
			return false, err
		}

		if hasChanged {
			return true, nil
		}
	}

	return false, nil
}

func (r *EndorserResult) Validate(state *endorserResultState) error {
	if len(state.Dependencies) != len(r.Dependencies) {
		return fmt.Errorf("number of dependencies does not match endorser result")
	}

	for i, dependency := range r.Dependencies {
		if err := dependency.Validate(&state.Dependencies[i]); err != nil {
			return fmt.Errorf("dependency state %v does not match dependency: %w", i, err)
		}
	}

	return nil
}

type endorserResultState struct {
	Dependencies []dependencyState
}

var parsedEndorserABI *abi.ABI

func useEndorserAbi() *abi.ABI {
	if parsedEndorserABI != nil {
		return parsedEndorserABI
	}

	parsed := ethcontract.MustParseABI(ENDORSER_ABI)
	parsedEndorserABI = &parsed
	return parsedEndorserABI
}

func IsOperationReady(ctx context.Context, provider *ethrpc.Provider, op *proto.Operation) (*EndorserResult, error) {
	endorserAddr := common.HexToAddress(op.Endorser)
	if endorserAddr == (common.Address{}) {
		return nil, fmt.Errorf("invalid endorser address")
	}

	ab := useEndorserAbi()
	endorser := ethcontract.NewContractCaller(endorserAddr, *ab, provider)

	entrypointAddr := common.HexToAddress(op.Entrypoint)
	calldataBytes := common.FromHex(op.CallData)
	endorserCalldataBytes := common.FromHex(op.EndorserCallData)
	gasLimitBigInt := new(big.Int).SetUint64(op.GasLimit)
	maxFeePerGasBigInt, ok := new(big.Int).SetString(op.MaxFeePerGas, 10)
	if !ok {
		return nil, fmt.Errorf("invalid max fee per gas")
	}
	priorityFeePerGas, ok := new(big.Int).SetString(op.PriorityFeePerGas, 10)
	if !ok {
		return nil, fmt.Errorf("invalid priority fee per gas")
	}
	feeToken := common.HexToAddress(op.FeeToken)

	calldata, err := endorser.Encode(
		"isOperationReady",
		entrypointAddr,
		calldataBytes,
		endorserCalldataBytes,
		gasLimitBigInt,
		maxFeePerGasBigInt,
		priorityFeePerGas,
		feeToken,
	)

	if err != nil {
		return nil, err
	}

	type Call struct {
		To   common.Address `json:"to"`
		Data string         `json:"data"`
	}

	endorserCall := &Call{
		To:   endorserAddr,
		Data: "0x" + common.Bytes2Hex(calldata),
	}

	var res string
	rpcCall := ethrpc.NewCallBuilder[string]("eth_call", nil, endorserCall, nil, nil)
	_, err = provider.Do(ctx, rpcCall.Into(&res))
	if err != nil {
		return nil, err
	}

	resBytes := common.FromHex(res)

	endorserResult := &EndorserResult{}

	dec1, err := ab.Methods["isOperationReady"].Outputs.Unpack(resBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to unpack result: %w", err)
	}

	// It must have 3 elements
	if len(dec1) != 3 {
		return nil, fmt.Errorf("invalid result length")
	}

	// First element must be a bool
	endorserResult.Readiness, ok = dec1[0].(bool)
	if !ok {
		return nil, fmt.Errorf("invalid readiness")
	}

	// Second element must be a struct
	dec2, ok := dec1[1].(struct {
		MaxNumber    *big.Int "json:\"maxNumber\""
		MaxTimestamp *big.Int "json:\"maxTimestamp\""
	})
	if !ok {
		return nil, fmt.Errorf("invalid block dependency")
	}

	endorserResult.BlockDependency = BlockDependency{
		MaxNumber:    dec2.MaxNumber,
		MaxTimestamp: dec2.MaxTimestamp,
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

	endorserResult.Dependencies = make([]Dependency, 0, len(dec3))
	for _, dep := range dec3 {
		dependency := Dependency{
			Addr:     dep.Addr,
			Balance:  dep.Balance,
			Code:     dep.Code,
			Nonce:    dep.Nonce,
			AllSlots: dep.AllSlots,
			Slots:    dep.Slots,
		}
		dependency.Constraint = make([]Constraint, 0, len(dep.Constraints))
		for _, c := range dep.Constraints {
			dependency.Constraint = append(dependency.Constraint, Constraint{
				Slot:     c.Slot,
				MinValue: c.MinValue,
				MaxValue: c.MaxValue,
			})
		}
		endorserResult.Dependencies = append(endorserResult.Dependencies, dependency)
	}

	return endorserResult, nil
}
