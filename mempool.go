package bundler

import (
	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/ethkit/ethrpc"
)

type TrackedOperation struct {
	proto.Operation
}

type Mempool struct {
	Provider *ethrpc.Provider
	MaxSize  uint

	FreshOperations []*proto.Operation
	Operations      []TrackedOperation
}

func NewMempool(cfg *config.MempoolConfig, provider *ethrpc.Provider) (*Mempool, error) {
	mp := &Mempool{
		Provider: provider,
		MaxSize:  cfg.Size,

		FreshOperations: []*proto.Operation{},
		Operations:      []TrackedOperation{},
	}

	return mp, nil
}
