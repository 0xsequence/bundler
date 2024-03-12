package types

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/0xsequence/bundler/contracts/gen/solabis/abiendorser"
	"github.com/0xsequence/bundler/ipfs"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/common/hexutil"
	"github.com/0xsequence/ethkit/go-ethereum/log"
	"github.com/0xsequence/go-sequence/lib/prototyp"
	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
)

type Operation struct {
	abiendorser.IEndorserOperation

	Endorser         common.Address `json:"endorser"`
	EndorserGasLimit *big.Int       `json:"endorserGasLimit"`
	ChainId          *big.Int       `json:"chainId"`
}

func NewOperation() *Operation {
	return &Operation{}
}

func (o *Operation) ToProto() *proto.Operation {
	pure := o.ToProtoPure()
	hash := o.Hash()
	pure.Hash = &hash
	return pure
}

func (o *Operation) ToProtoPure() *proto.Operation {
	return &proto.Operation{
		Entrypoint:             prototyp.ToHash(o.Entrypoint),
		Data:                   prototyp.HashFromBytes(o.Data),
		GasLimit:               prototyp.ToBigInt(o.GasLimit),
		FeeToken:               prototyp.ToHash(o.FeeToken),
		Endorser:               prototyp.ToHash(o.Endorser),
		EndorserCallData:       prototyp.HashFromBytes(o.EndorserCallData),
		EndorserGasLimit:       prototyp.ToBigInt(o.EndorserGasLimit),
		MaxFeePerGas:           prototyp.ToBigInt(o.MaxFeePerGas),
		MaxPriorityFeePerGas:   prototyp.ToBigInt(o.MaxPriorityFeePerGas),
		FeeScalingFactor:       prototyp.ToBigInt(o.FeeScalingFactor),
		FeeNormalizationFactor: prototyp.ToBigInt(o.FeeNormalizationFactor),
		HasUntrustedContext:    o.HasUntrustedContext,
		ChainID:                prototyp.ToBigInt(o.ChainId),
	}
}

func NewOperationFromProto(op *proto.Operation) (*Operation, error) {
	if !op.Entrypoint.IsValidAddress() {
		return nil, fmt.Errorf("invalid entrypoint address \"%v\"", op.Entrypoint)
	}
	entrypoint := op.Entrypoint.ToAddress()

	calldata, err := hexutil.Decode(op.Data.String())
	if err != nil {
		return nil, fmt.Errorf("invalid calldata hex string: %w", err)
	}

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

	endorserCalldata, err := hexutil.Decode(op.EndorserCallData.String())
	if err != nil {
		return nil, fmt.Errorf("invalid endorser calldata hex string: %w", err)
	}

	if op.EndorserGasLimit.Int().Sign() <= 0 {
		return nil, fmt.Errorf("invalid endorser gas limit %v", op.EndorserGasLimit)
	}

	return &Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			Entrypoint:             entrypoint,
			Data:                   calldata,
			GasLimit:               op.GasLimit.Int(),
			FeeToken:               feeToken,
			EndorserCallData:       endorserCalldata,
			MaxFeePerGas:           op.MaxFeePerGas.Int(),
			MaxPriorityFeePerGas:   op.MaxPriorityFeePerGas.Int(),
			FeeScalingFactor:       op.FeeScalingFactor.Int(),
			FeeNormalizationFactor: op.FeeNormalizationFactor.Int(),
			HasUntrustedContext:    op.HasUntrustedContext,
		},
		Endorser:         endorser,
		EndorserGasLimit: op.EndorserGasLimit.Int(),
		ChainId:          op.ChainID.Int(),
	}, nil
}

func (op *Operation) NativePayment() bool {
	return op.FeeToken == common.Address{}
}

func (op *Operation) Hash() string {
	// Convert to json
	jsonData, err := json.Marshal(op.ToProtoPure())
	if err != nil {
		return ""
	}

	// Normalize
	jsonData, err = jsoncanonicalizer.Transform(jsonData)
	if err != nil {
		return ""
	}

	res, err := ipfs.Cid(jsonData)
	if err != nil {
		log.Warn("failed to create CID", "error", err)
	}

	return res
}

func (op *Operation) ReportToIPFS(ip ipfs.Interface) error {
	// Convert to json
	jsonData, err := json.Marshal(op.ToProtoPure())
	if err != nil {
		return err
	}

	// Normalize
	jsonData, err = jsoncanonicalizer.Transform(jsonData)
	if err != nil {
		return fmt.Errorf("unable to normalize operation json: %w", err)
	}

	cid, err := ipfs.Cid(jsonData)
	if err != nil {
		return err
	}

	res, err := ip.Report(jsonData)
	if err != nil {
		return err
	}

	if res != cid {
		return fmt.Errorf("CID mismatch %s != %s", res, cid)
	}

	return nil
}
