package rpc

import (
	"fmt"

	"github.com/0xsequence/ethkit/ethwallet"
)

func SetupWallet(mnemonic string, accountIndex uint32) (*ethwallet.Wallet, error) {
	wallet, err := ethwallet.NewWalletFromMnemonic(mnemonic)
	if err != nil {
		return nil, fmt.Errorf("unable to create wallet from mnemonic: %w", err)
	}

	_, err = wallet.SelfDeriveAccountIndex(accountIndex)
	if err != nil {
		return nil, fmt.Errorf("unable to derive account %v: %w", accountIndex)
	}

	return wallet, nil
}
