package interfaces

import (
	"context"

	"github.com/0xsequence/ethkit/ethtxn"
	"github.com/0xsequence/ethkit/ethwallet"
	"github.com/0xsequence/ethkit/go-ethereum/common"

	ethtypes "github.com/0xsequence/ethkit/go-ethereum/core/types"
)

type Wallet interface {
	Address() common.Address
	GetNonce(ctx context.Context) (uint64, error)
	NewTransaction(ctx context.Context, txnRequest *ethtxn.TransactionRequest) (*ethtypes.Transaction, error)
	SendTransaction(ctx context.Context, t *ethtypes.Transaction) (*ethtypes.Transaction, ethtxn.WaitReceipt, error)
}

var _ Wallet = &ethwallet.Wallet{}
