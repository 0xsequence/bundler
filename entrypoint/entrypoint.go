package entrypoint

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

const ENTRYPOINT_ABI = `[{"type":"function","name":"safeExecute","inputs":[{"name":"_entrypoint","type":"address","internalType":"address"},{"name":"_data","type":"bytes","internalType":"bytes"},{"name":"_gasLimit","type":"uint256","internalType":"uint256"},{"name":"_maxFeePerGas","type":"uint256","internalType":"uint256"},{"name":"_maxPriorityFeePerGas","type":"uint256","internalType":"uint256"},{"name":"_feeToken","type":"address","internalType":"address"},{"name":"_calldataGas","type":"uint256","internalType":"uint256"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"simulateOperation","inputs":[{"name":"_entrypoint","type":"address","internalType":"address"},{"name":"_data","type":"bytes","internalType":"bytes"},{"name":"_endorserCallData","type":"bytes","internalType":"bytes"},{"name":"_gasLimit","type":"uint256","internalType":"uint256"},{"name":"_maxFeePerGas","type":"uint256","internalType":"uint256"},{"name":"_maxPriorityFeePerGas","type":"uint256","internalType":"uint256"},{"name":"_feeToken","type":"address","internalType":"address"},{"name":"_endorser","type":"address","internalType":"address"},{"name":"_calldataGas","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"result","type":"tuple","internalType":"struct BundlerEntrypoint.SimulationResult","components":[{"name":"paid","type":"bool","internalType":"bool"},{"name":"lied","type":"bool","internalType":"bool"}]}],"stateMutability":"nonpayable"},{"type":"error","name":"BundlerExecutionFailed","inputs":[]},{"type":"error","name":"BundlerUnderpaid","inputs":[{"name":"_paid","type":"uint256","internalType":"uint256"},{"name":"_expected","type":"uint256","internalType":"uint256"}]}]`

var parsedEntrypointABI *abi.ABI

func useEntrypointABI() *abi.ABI {
	if parsedEntrypointABI != nil {
		return parsedEntrypointABI
	}

	parsed := ethcontract.MustParseABI(ENTRYPOINT_ABI)
	parsedEntrypointABI = &parsed
	return parsedEntrypointABI
}

func Simulate(ctx context.Context, topEntrypointAddr common.Address, provider *ethrpc.Provider, op *proto.Operation) (bool, bool, error) {
	ab := useEntrypointABI()
	topEntrypoint := ethcontract.NewContractCaller(topEntrypointAddr, *ab, provider)

	entrypointAddr := common.HexToAddress(op.Entrypoint)
	endorserAddr := common.HexToAddress(op.Endorser)
	calldataBytes := common.FromHex(op.CallData)
	endorserCalldataBytes := common.FromHex(op.EndorserCallData)
	gasLimitBigInt := new(big.Int).SetUint64(op.GasLimit)
	maxFeePerGasBigInt, ok := new(big.Int).SetString(op.MaxFeePerGas, 10)
	if !ok {
		return false, false, fmt.Errorf("invalid max fee per gas")
	}
	priorityFeePerGas, ok := new(big.Int).SetString(op.PriorityFeePerGas, 10)
	if !ok {
		return false, false, fmt.Errorf("invalid priority fee per gas")
	}
	feeToken := common.HexToAddress(op.FeeToken)

	calldata, err := topEntrypoint.Encode(
		"simulateOperation",
		entrypointAddr,
		calldataBytes,
		endorserCalldataBytes,
		gasLimitBigInt,
		maxFeePerGasBigInt,
		priorityFeePerGas,
		feeToken,
		endorserAddr,
		big.NewInt(0), // Better compute calldata gas
	)

	if err != nil {
		return false, false, err
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
		return false, false, err
	}

	resBytes := common.FromHex(res)

	dec1, err := ab.Methods["isOperationReady"].Outputs.Unpack(resBytes)
	if err != nil {
		return false, false, fmt.Errorf("unable to unpack result: %w", err)
	}

	dec2, ok := dec1[2].([]struct {
		Paid bool "json:\"paid\""
		Lied bool "json:\"lied\""
	})

	if !ok {
		return false, false, fmt.Errorf("unable to unpack result")
	}

	return dec2[0].Paid, dec2[0].Lied, nil
}
