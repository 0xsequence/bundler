package endorser

import (
	"context"
	"fmt"
	"math/big"

	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/ethkit/ethcontract"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi"
	"github.com/0xsequence/ethkit/go-ethereum/common"
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
