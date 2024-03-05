package sender

import (
	"math/big"

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
	SimulateOperation(opts *bind.CallOpts, _entrypoint common.Address, _data []byte, _endorserCallData []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _hasUntrustedContext bool, _endorser common.Address) (abivalidator.OperationValidatorSimulationResult, error)
}

var _ ValidatorInterface = &abivalidator.OperationValidator{}

type Interface interface {
	Run(ctx context.Context)
}
