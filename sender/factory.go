package sender

import (
	"fmt"

	"github.com/0xsequence/bundler/lib/interfaces"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/ethwallet"
)

type MnemonicWalletFactory struct {
	provider *ethrpc.Provider
	mnemonic string
}

func NewMnemonicWalletFactory(provider *ethrpc.Provider, mnemonic string) *MnemonicWalletFactory {
	return &MnemonicWalletFactory{
		provider: provider,
		mnemonic: mnemonic,
	}
}

func (f *MnemonicWalletFactory) GetWallet(i int) (interfaces.Wallet, error) {
	wallet, err := ethwallet.NewWalletFromMnemonic(f.mnemonic)
	if err != nil {
		return nil, fmt.Errorf("unable to create wallet from mnemonic: %w", err)
	}

	_, err = wallet.SelfDeriveAccountIndex(uint32(i))
	if err != nil {
		return nil, fmt.Errorf("unable to derive account %v: %w", i, err)
	}

	wallet.SetProvider(f.provider)

	return wallet, nil
}

var _ WalletFactory = &MnemonicWalletFactory{}
