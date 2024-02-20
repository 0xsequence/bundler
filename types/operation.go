package types

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/common/hexutil"
	"github.com/0xsequence/ethkit/go-ethereum/crypto"
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

// TODO: Change this
func (op *Operation) Digest() common.Hash {
	o := op.ToProto()

	var gasLimit [8]byte
	var endorserGasLimit [8]byte

	binary.LittleEndian.PutUint64(gasLimit[:], o.GasLimit)
	binary.LittleEndian.PutUint64(endorserGasLimit[:], o.EndorserGasLimit)

	maxFeePerGas, ok := new(big.Int).SetString(o.MaxFeePerGas, 0)
	if !ok {
		maxFeePerGas = big.NewInt(0)
	}

	priorityFeePerGas, ok := new(big.Int).SetString(o.PriorityFeePerGas, 0)
	if !ok {
		priorityFeePerGas = big.NewInt(0)
	}

	baseFeeScalingFactor, ok := new(big.Int).SetString(o.BaseFeeScalingFactor, 0)
	if !ok {
		baseFeeScalingFactor = big.NewInt(0)
	}

	baseFeeNormalizationFactor, ok := new(big.Int).SetString(o.BaseFeeNormalizationFactor, 0)
	if !ok {
		baseFeeNormalizationFactor = big.NewInt(0)
	}

	hasUntrustedContext := []byte{0}
	if o.HasUntrustedContext {
		hasUntrustedContext = []byte{1}
	}

	return crypto.Keccak256Hash(
		common.HexToAddress(o.Entrypoint).Bytes(),
		hexutil.MustDecode(o.CallData),
		gasLimit[:],
		common.HexToAddress(o.FeeToken).Bytes(),
		common.HexToAddress(o.Endorser).Bytes(),
		hexutil.MustDecode(o.EndorserCallData),
		endorserGasLimit[:],
		maxFeePerGas.Bytes(),
		priorityFeePerGas.Bytes(),
		baseFeeScalingFactor.Bytes(),
		baseFeeNormalizationFactor.Bytes(),
		hasUntrustedContext,
	)
}
