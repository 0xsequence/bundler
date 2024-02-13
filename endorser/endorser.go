package endorser

import (
	"context"
	"fmt"
	"math/big"

	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/ethkit/ethcoder"
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
	Addr       *common.Address
	Balance    bool
	Code       bool
	Nonce      bool
	AllSlots   bool
	Slots      [][32]byte
	Constraint []Constraint
}

type EndorserResult struct {
	Readiness       bool
	BlockDependency BlockDependency
	Dependencies    []Dependency
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

	endorser := ethcontract.NewContractCaller(endorserAddr, *useEndorserAbi(), provider)
	calldata, err := endorser.Encode(
		"isOperationReady",
		op.Entrypoint,
		op.CallData,
		op.EndorserCallData,
		op.GasLimit,
		op.MaxFeePerGas,
		op.PriorityFeePerGas,
		op.FeeToken,
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

	var readiness bool
	var blockDependency BlockDependency
	var dependencies []Dependency

	err = ethcoder.AbiDecoder(
		[]string{"bool", "tuple(uint256,uint256)", "tuple(address,bool,bool,bool,bool,bytes32[],tuple(bytes32,bytes32,bytes32)[])[]"},
		resBytes,
		[]interface{}{&readiness, &blockDependency, &dependencies},
	)

	if err != nil {
		return nil, err
	}

	return &EndorserResult{
		Readiness:       readiness,
		BlockDependency: blockDependency,
		Dependencies:    dependencies,
	}, nil
}
