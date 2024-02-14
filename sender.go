package bundler

import "github.com/0xsequence/ethkit/ethwallet"

type Sender struct {
	ID uint32

	Wallet  *ethwallet.Wallet
	Mempool *Mempool
}

func NewSender(id uint32, wallet *ethwallet.Wallet, mempool *Mempool) *Sender {
	return &Sender{
		ID:      id,
		Wallet:  wallet,
		Mempool: mempool,
	}
}
