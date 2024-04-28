package sender

import (
	"github.com/0xsequence/bundler/lib/interfaces"
	"golang.org/x/net/context"
)

type Interface interface {
	Run(ctx context.Context)
}

type WalletFactory interface {
	GetWallet(i int) (interfaces.Wallet, error)
}
