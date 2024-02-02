package bundler

import (
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/common/hexutil"
)

type Operation struct {
	Entrypoint common.Address `json:"entrypoint"`
	Calldata   hexutil.Bytes  `json:"calldata"`
	TS         uint64         `json:"ts"`
}
