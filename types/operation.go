package types

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/log"
	"github.com/0xsequence/go-sequence/lib/prototyp"
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
		Entrypoint:                 prototyp.ToHash(o.Entrypoint),
		CallData:                   prototyp.HashFromBytes(o.Calldata),
		GasLimit:                   prototyp.ToBigInt(o.GasLimit),
		FeeToken:                   prototyp.ToHash(o.FeeToken),
		Endorser:                   prototyp.ToHash(o.Endorser),
		EndorserCallData:           prototyp.HashFromBytes(o.EndorserCallData),
		EndorserGasLimit:           prototyp.ToBigInt(o.EndorserGasLimit),
		MaxFeePerGas:               prototyp.ToBigInt(o.MaxFeePerGas),
		PriorityFeePerGas:          prototyp.ToBigInt(o.PriorityFeePerGas),
		BaseFeeScalingFactor:       prototyp.ToBigInt(o.BaseFeeScalingFactor),
		BaseFeeNormalizationFactor: prototyp.ToBigInt(o.BaseFeeNormalizationFactor),
		HasUntrustedContext:        o.HasUntrustedContext,
	}
}

func NewOperationFromProto(op *proto.Operation) (*Operation, error) {
	if !op.Entrypoint.IsValidAddress() {
		return nil, fmt.Errorf("invalid entrypoint address \"%v\"", op.Entrypoint)
	}
	entrypoint := op.Entrypoint.ToAddress()

	if op.GasLimit.Int().Sign() <= 0 {
		return nil, fmt.Errorf("invalid gas limit %v", op.GasLimit)
	}

	if !op.FeeToken.IsValidAddress() {
		return nil, fmt.Errorf("invalid fee token address \"%v\"", op.FeeToken)
	}
	feeToken := op.FeeToken.ToAddress()

	if !op.Endorser.IsValidAddress() {
		return nil, fmt.Errorf("invalid endorser address \"%v\"", op.Endorser)
	}
	endorser := op.Endorser.ToAddress()

	if op.EndorserGasLimit.Int().Sign() <= 0 {
		return nil, fmt.Errorf("invalid endorser gas limit %v", op.EndorserGasLimit)
	}

	return &Operation{
		Entrypoint:                 entrypoint,
		Calldata:                   op.CallData.Bytes(),
		GasLimit:                   op.GasLimit.Int(),
		FeeToken:                   feeToken,
		Endorser:                   endorser,
		EndorserCallData:           op.EndorserCallData.Bytes(),
		EndorserGasLimit:           op.EndorserGasLimit.Int(),
		MaxFeePerGas:               op.MaxFeePerGas.Int(),
		PriorityFeePerGas:          op.PriorityFeePerGas.Int(),
		BaseFeeScalingFactor:       op.BaseFeeScalingFactor.Int(),
		BaseFeeNormalizationFactor: op.BaseFeeScalingFactor.Int(),
		HasUntrustedContext:        op.HasUntrustedContext,
	}, nil
}

func (op *Operation) Digest() string {
	// Convert to json
	jsonData, err := json.Marshal(op.ToProto())
	if err != nil {
		return ""
	}

	res, err := Cid(jsonData)
	if err != nil {
		log.Warn("failed to create CID", "error", err)
	}

	return res
}

func (op *Operation) ReportToIPFS(url string) error {
	// Convert to json
	jsonData, err := json.Marshal(op.ToProto())
	if err != nil {
		return err
	}

	cid, err := Cid(jsonData)
	if err != nil {
		return err
	}

	res, err := ReportToIPFS(url, jsonData)
	if err != nil {
		return err
	}

	if res != cid {
		return fmt.Errorf("CID mismatch %s != %s", res, cid)
	}

	return nil
}
