package rpc

import (
	"fmt"

	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/ethwallet"
)

func SetupWallet(mnemonic string, accountIndex uint32, provider *ethrpc.Provider) (*ethwallet.Wallet, error) {
	wallet, err := ethwallet.NewWalletFromMnemonic(mnemonic)
	if err != nil {
		return nil, fmt.Errorf("unable to create wallet from mnemonic: %w", err)
	}

	_, err = wallet.SelfDeriveAccountIndex(accountIndex)
	if err != nil {
		return nil, fmt.Errorf("unable to derive account %v: %w", accountIndex, err)
	}

	wallet.SetProvider(provider)

	return wallet, nil
}
