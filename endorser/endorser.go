package endorser

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/0xsequence/bundler/contracts/gen/solabis/abiendorser"
	"github.com/0xsequence/bundler/types"
	"github.com/0xsequence/ethkit/ethcontract"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi"
	"github.com/0xsequence/ethkit/go-ethereum/common"
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

func IsOperationReady(ctx context.Context, provider *ethrpc.Provider, op *types.Operation) (*EndorserResult, error) {
	ab := useEndorserAbi()
	endorser := ethcontract.NewContractCaller(op.Endorser, *ab, provider)

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
	_, err = provider.Do(ctx, rpcCall.Into(&res))
	if err != nil {
		if strings.Contains(err.Error(), "execution reverted") {
			// TODO: Add the reason, as it may be useful
			// for someone adding a new op using an RPC call
			// return &EndorserResult{
			// 	Readiness: false,
			// }, nil
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
