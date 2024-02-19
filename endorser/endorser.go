package endorser

import (
	"context"
	"fmt"
	"math/big"
	"strings"

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
			},
			{
				"internalType": "uint256",
				"name": "_baseFeeScalingFactor",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "_baseFeeNormalizationFactor",
				"type": "uint256"
			},
			{
				"internalType": "bool",
				"name": "_hasUntrustedContext",
				"type": "bool"
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
						"internalType": "bool",
						"name": "basefee",
						"type": "bool"
					},
					{
						"internalType": "bool",
						"name": "blobbasefee",
						"type": "bool"
					},
					{
						"internalType": "bool",
						"name": "chainid",
						"type": "bool"
					},
					{
						"internalType": "bool",
						"name": "coinbase",
						"type": "bool"
					},
					{
						"internalType": "bool",
						"name": "difficulty",
						"type": "bool"
					},
					{
						"internalType": "bool",
						"name": "gasLimit",
						"type": "bool"
					},
					{
						"internalType": "bool",
						"name": "number",
						"type": "bool"
					},
					{
						"internalType": "bool",
						"name": "timestamp",
						"type": "bool"
					},
					{
						"internalType": "bool",
						"name": "txOrigin",
						"type": "bool"
					},
					{
						"internalType": "bool",
						"name": "txGasPrice",
						"type": "bool"
					},
					{
						"internalType": "uint256",
						"name": "maxBlockNumber",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "maxBlockTimestamp",
						"type": "uint256"
					}
				],
				"internalType": "struct Endorser.GlobalDependency",
				"name": "globalDependency",
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

type GlobalDependency struct {
	BaseFee           bool
	BlobBaseFee       bool
	ChainID           bool
	Coinbase          bool
	Difficulty        bool
	BlockGasLimit     bool
	BlockNumber       bool
	BlockTimestamp    bool
	TxOrigin          bool
	TxGasPrice        bool
	MaxBlockNumber    *big.Int
	MaxBlockTimestamp *big.Int
}

type Constraint struct {
	Slot     [32]byte
	MinValue [32]byte
	MaxValue [32]byte
}

type Dependency struct {
	Addr        common.Address
	Balance     bool
	Code        bool
	Nonce       bool
	AllSlots    bool
	Slots       [][32]byte
	Constraints []Constraint
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
	baseFeeScalingFactor, ok := new(big.Int).SetString(op.BaseFeeScalingFactor, 10)
	if !ok {
		return nil, fmt.Errorf("invalid base fee scaling factor")
	}
	baseFeeNormalizationFactor, ok := new(big.Int).SetString(op.BaseFeeNormalizationFactor, 10)
	if !ok {
		return nil, fmt.Errorf("invalid base fee normalization factor")
	}

	calldata, err := endorser.Encode(
		"isOperationReady",
		entrypointAddr,
		calldataBytes,
		endorserCalldataBytes,
		gasLimitBigInt,
		maxFeePerGasBigInt,
		priorityFeePerGas,
		feeToken,
		baseFeeScalingFactor,
		baseFeeNormalizationFactor,
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
		To:   endorserAddr,
		Data: "0x" + common.Bytes2Hex(calldata),
	}

	var res string
	rpcCall := ethrpc.NewCallBuilder[string]("eth_call", nil, endorserCall, nil, nil)
	_, err = provider.Do(ctx, rpcCall.Into(&res))
	if err != nil {
		if strings.Contains(err.Error(), "execution reverted") {
			// TODO: Add the reason, as it may be useful
			// for someone adding a new op using an RPC call
			return &EndorserResult{
				Readiness: false,
			}, nil
		}

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

	endorserResult.GlobalDependency = GlobalDependency{
		BaseFee:           dec2.Basefee,
		BlobBaseFee:       dec2.Blobbasefee,
		ChainID:           dec2.Chainid,
		Coinbase:          dec2.Coinbase,
		Difficulty:        dec2.Difficulty,
		BlockGasLimit:     dec2.GasLimit,
		BlockNumber:       dec2.Number,
		BlockTimestamp:    dec2.Timestamp,
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
		dependency.Constraints = make([]Constraint, 0, len(dep.Constraints))
		for _, c := range dep.Constraints {
			dependency.Constraints = append(dependency.Constraints, Constraint{
				Slot:     c.Slot,
				MinValue: c.MinValue,
				MaxValue: c.MaxValue,
			})
		}
		endorserResult.Dependencies = append(endorserResult.Dependencies, dependency)
	}

	return endorserResult, nil
}
