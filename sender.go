package bundler

import "github.com/0xsequence/ethkit/ethwallet"

type Sender struct {
	ID     uint32
	Wallet *ethwallet.Wallet
}
