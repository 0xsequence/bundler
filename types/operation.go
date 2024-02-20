package types

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/log"
)

type Operation struct {
	Entrypoint                 common.Address `json:"entrypoint"`
	Calldata                   []byte         `json:"callData"`
	GasLimit                   *big.Int       `json:"gasLimit"`
	FeeToken                   common.Address `json:"feeToken"`
	Endorser                   common.Address `json:"endorser"`
	EndorserCallData           []byte         `json:"endorserCallData"`
	EndorserGasLimit           *big.Int       `json:"endorserGasLimit"`
	MaxFeePerGas               *big.Int       `json:"maxFeePerGas"`
	PriorityFeePerGas          *big.Int       `json:"priorityFeePerGas"`
	BaseFeeScalingFactor       *big.Int       `json:"baseFeeScalingFactor"`
	BaseFeeNormalizationFactor *big.Int       `json:"baseFeeNormalizationFactor"`
	HasUntrustedContext        bool           `json:"hasUntrustedContext"`
}

func NewOperation() *Operation {
	return &Operation{}
}

func (o *Operation) ToProto() *proto.Operation {
	return &proto.Operation{
		Entrypoint:                 o.Entrypoint.String(),
		CallData:                   "0x" + common.Bytes2Hex(o.Calldata),
		GasLimit:                   o.GasLimit.Uint64(),
		FeeToken:                   o.FeeToken.String(),
		Endorser:                   o.Endorser.String(),
		EndorserCallData:           "0x" + common.Bytes2Hex(o.EndorserCallData),
		EndorserGasLimit:           o.EndorserGasLimit.Uint64(),
		MaxFeePerGas:               o.MaxFeePerGas.String(),
		PriorityFeePerGas:          o.PriorityFeePerGas.String(),
		BaseFeeScalingFactor:       o.BaseFeeScalingFactor.String(),
		BaseFeeNormalizationFactor: o.BaseFeeNormalizationFactor.String(),
		HasUntrustedContext:        o.HasUntrustedContext,
	}
}

func (o *Operation) FromProto(op *proto.Operation) (*Operation, error) {
	endorser := common.HexToAddress(op.Endorser)
	if endorser == (common.Address{}) {
		return nil, fmt.Errorf("invalid endorser address")
	}

	entrypoint := common.HexToAddress(op.Entrypoint)
	if entrypoint == (common.Address{}) {
		return nil, fmt.Errorf("invalid entrypoint address")
	}

	calldata, err := FromHex(op.CallData)
	if err != nil {
		return nil, err
	}

	endorserCalldata, err := FromHex(op.EndorserCallData)
	if err != nil {
		return nil, err
	}

	if op.GasLimit <= 0 {
		return nil, fmt.Errorf("invalid gas limit")
	}
	gasLimit := new(big.Int).SetUint64(op.GasLimit)

	if op.EndorserGasLimit <= 0 {
		return nil, fmt.Errorf("invalid endorser gas limit")
	}
	endorserGasLimit := new(big.Int).SetUint64(op.EndorserGasLimit)

	maxFeePerGas, err := HexToBigInt(op.MaxFeePerGas)
	if err != nil {
		return nil, err
	}

	priorityFeePerGas, err := HexToBigInt(op.PriorityFeePerGas)
	if err != nil {
		return nil, err
	}

	baseFeeScalingFactor, err := HexToBigInt(op.BaseFeeScalingFactor)
	if err != nil {
		return nil, err
	}

	baseFeeNormalizationFactor, err := HexToBigInt(op.BaseFeeNormalizationFactor)
	if err != nil {
		return nil, err
	}

	o.Entrypoint = entrypoint
	o.Calldata = calldata
	o.GasLimit = gasLimit
	o.FeeToken = common.HexToAddress(op.FeeToken)
	o.Endorser = endorser
	o.EndorserCallData = endorserCalldata
	o.EndorserGasLimit = endorserGasLimit
	o.MaxFeePerGas = maxFeePerGas
	o.PriorityFeePerGas = priorityFeePerGas
	o.BaseFeeScalingFactor = baseFeeScalingFactor
	o.BaseFeeNormalizationFactor = baseFeeNormalizationFactor
	o.HasUntrustedContext = op.HasUntrustedContext

	return o, nil
}

func (op *Operation) Digest() string {
	o := op.ToProto()

	// Convert to json
	jsonData, err := json.Marshal(o)
	if err != nil {
		return ""
	}

	// return base58.Encode(mhash)
	res, err := Cid(jsonData)
	if err != nil {
		log.Warn("failed to create CID", "error", err)
	}

	return res
}
