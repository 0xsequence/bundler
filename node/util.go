package node

import (
	"github.com/0xsequence/ethkit/ethwallet"
	"github.com/0xsequence/ethkit/go-ethereum/accounts"
	"github.com/0xsequence/ethkit/go-ethereum/common/hexutil"
)

func setupWallet(privateKeyHex string, derivationPath string) (*ethwallet.Wallet, error) {
	privKey, err := hexutil.Decode(privateKeyHex)
	if err != nil {
		return nil, err
	}

	derivPath := accounts.DefaultBaseDerivationPath
	if derivationPath != "" {
		derivPath, err = ethwallet.ParseDerivationPath(derivationPath)
		if err != nil {
			return nil, err
		}
	}

	hdnode, err := ethwallet.NewHDNodeFromEntropy(privKey, &derivPath)
	if err != nil {
		return nil, err
	}
	// fmt.Println("hdnode", hdnode.Address().String(), hdnode.DerivationPath().String())

	// Create ethereum HD wallet used by the txn senders.
	wallet, err := ethwallet.NewWalletFromHDNode(hdnode)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}
