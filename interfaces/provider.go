package interfaces

import (
	"context"
	"math/big"

	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	ethtypes "github.com/0xsequence/ethkit/go-ethereum/core/types"
)

type Provider interface {
	BlockByHash(ctx context.Context, hash common.Hash) (*ethtypes.Block, error)
	CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error)
	BalanceAt(ctx context.Context, account common.Address, blockNum *big.Int) (*big.Int, error)
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNum *big.Int) ([]byte, error)
	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
}

var _ Provider = &ethrpc.Provider{}
