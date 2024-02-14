package bundler

import (
	"context"
	"math/big"

	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/ethwallet"
	"github.com/0xsequence/ethkit/go-ethereum/common"
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
}

func NewSender(id uint32, wallet *ethwallet.Wallet, mempool *Mempool, provider *ethrpc.Provider) *Sender {
	chainID, err := provider.ChainID(context.TODO())
	if err != nil {
	}

	return &Sender{
		ID:       id,
		Wallet:   wallet,
		Mempool:  mempool,
		Provider: provider,
		ChainID:  chainID,
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

		for _, op := range ops {
			state, err := op.EndorserResult.State(ctx, s.Provider)
			if err != nil {
			}

			hasChanged, err := op.EndorserResult.HasChanged(op.EndorserResultState, state)
			if err != nil {
			}

			if hasChanged {
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
