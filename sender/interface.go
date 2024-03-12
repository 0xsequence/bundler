package sender

import (
	"github.com/0xsequence/bundler/contracts/gen/solabis/abivalidator"
	"github.com/0xsequence/ethkit/ethtxn"
	"github.com/0xsequence/ethkit/ethwallet"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi/bind"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	ethtypes "github.com/0xsequence/ethkit/go-ethereum/core/types"
	"golang.org/x/net/context"
)

type WalletInterface interface {
	Address() common.Address
	GetNonce(ctx context.Context) (uint64, error)
	NewTransaction(ctx context.Context, txnRequest *ethtxn.TransactionRequest) (*ethtypes.Transaction, error)
	SendTransaction(ctx context.Context, t *ethtypes.Transaction) (*ethtypes.Transaction, ethtxn.WaitReceipt, error)
}

var _ WalletInterface = &ethwallet.Wallet{}

type ValidatorInterface interface {
	SimulateOperation(opts *bind.CallOpts, _endorser common.Address, _op abivalidator.IEndorserOperation) (abivalidator.OperationValidatorSimulationResult, error)
}

var _ ValidatorInterface = &abivalidator.OperationValidator{}

type Interface interface {
	Run(ctx context.Context)
}
