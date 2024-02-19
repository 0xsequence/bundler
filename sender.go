package bundler

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/0xsequence/bundler/contracts/gen/operationvalidator"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/ethwallet"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi/bind"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/common/hexutil"
	"github.com/0xsequence/ethkit/go-ethereum/core/types"
	"github.com/davecgh/go-spew/spew"
)

const BatchSize = 10

type Sender struct {
	ID uint32

	Wallet   *ethwallet.Wallet
	Mempool  *Mempool
	Provider *ethrpc.Provider
	ChainID  *big.Int

	executor  *operationvalidator.OperationValidator
	simulator *operationvalidator.OperationValidatorSimulator
}

func NewSender(id uint32, wallet *ethwallet.Wallet, mempool *Mempool, provider *ethrpc.Provider, executor *operationvalidator.OperationValidator, simulator *operationvalidator.OperationValidatorSimulator) *Sender {
	chainID, err := provider.ChainID(context.TODO())
	if err != nil {
	}

	return &Sender{
		ID:       id,
		Wallet:   wallet,
		Mempool:  mempool,
		Provider: provider,
		ChainID:  chainID,

		executor:  executor,
		simulator: simulator,
	}
}

func (s *Sender) Run(ctx context.Context) {
	var execute, discard []*TrackedOperation

	for ctx.Err() == nil {
		ops := s.Mempool.ReserveOps(ctx, func(to []*TrackedOperation) []*TrackedOperation {
			if BatchSize < len(to) {
				return to[:BatchSize]
			} else {
				return to
			}
		})

		if len(ops) == 0 {
			time.Sleep(time.Second)
			continue
		}

		for _, op := range ops {
			result, err := endorser.IsOperationReady(ctx, s.Provider, &op.Operation)
			if err != nil {
			}
			if op.EndorserResult.Readiness {
				execute = append(execute, op)
			} else {
				discard = append(discard, op)
			}

			op.EndorserResult = result
		}

		s.Mempool.DiscardOps(ctx, discard)
		discard = nil

		for _, op := range execute {
			to := common.HexToAddress(op.Entrypoint)
			data := common.Hex2Bytes(op.CallData)

			tx, err := s.Wallet.SignTx(types.NewTx(&types.DynamicFeeTx{
				To:   &to,
				Data: data,
			}), s.ChainID)
			if err != nil {
			}

			_, wait, err := s.Wallet.SendTransaction(ctx, tx)
			if err != nil {
			}

			receipt, err := wait(ctx)
			if err != nil {
			}

			spew.Dump(receipt)
		}

		execute = nil
	}
}

func (s *Sender) simulateOperation(ctx context.Context, op *proto.Operation) (paid bool, lied bool, err error) {
	gasLimit := new(big.Int).SetUint64(op.GasLimit)

	maxFeePerGas, ok := new(big.Int).SetString(op.MaxFeePerGas, 0)
	if !ok {
		return false, false, fmt.Errorf("maxFeePerGas \"%v\" is not a number", op.MaxFeePerGas)
	}

	maxPriorityFeePerGas, ok := new(big.Int).SetString(op.PriorityFeePerGas, 0)
	if !ok {
		return false, false, fmt.Errorf("maxPriorityFeePerGas \"%v\" is not a number", op.PriorityFeePerGas)
	}

	baseFeeScalingFactor, ok := new(big.Int).SetString(op.BaseFeeScalingFactor, 0)
	if !ok {
		return false, false, fmt.Errorf("baseFeeScalingFactor \"%v\" is not a number", op.BaseFeeScalingFactor)
	}

	baseFeeNormalizationFactor, ok := new(big.Int).SetString(op.BaseFeeNormalizationFactor, 0)
	if !ok {
		return false, false, fmt.Errorf("baseFeeNormalizationFactor \"%v\" is not a number", op.BaseFeeNormalizationFactor)
	}

	// TODO: compute this properly later
	var callDataGasUsage int64
	for _, b := range hexutil.MustDecode(op.CallData) {
		switch b {
		case 0:
		callDataGasUsage += 4
		default:
		callDataGasUsage += 16
		}
	}

	result, err := s.simulator.SimulateOperation(&bind.CallOpts{Context: ctx}, common.HexToAddress(op.Entrypoint), hexutil.MustDecode(op.CallData), hexutil.MustDecode(op.EndorserCallData), gasLimit, maxFeePerGas, maxPriorityFeePerGas, common.HexToAddress(op.FeeToken), baseFeeScalingFactor, baseFeeNormalizationFactor, op.HasUntrustedContext, common.HexToAddress(op.Endorser), big.NewInt(callDataGasUsage))
	if err != nil {
		return false, false, fmt.Errorf("unable to call simulateOperation: %w", err)
	}

	// if result.lied is true and the constraints are all satisfied, then we should ban the endorser
	// if result.paid is false, then we should drop the transaction

	// TODO: only return lied if the constraints are all satisfied
	if result.Lied {
		return false, true, nil
	}

	if result.Paid {
		return true, false, nil
	}

	return false, false, nil
}
