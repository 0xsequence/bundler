package bundler

import "github.com/0xsequence/bundler/proto"

type Mempool struct {
	BundleOps    []proto.Operation `json:"bundleOps"`
	SingletonOps []proto.Operation `json:"singletonOps"`
}

func NewMempool() (*Mempool, error) {
	mp := &Mempool{
		BundleOps:    []proto.Operation{},
		SingletonOps: []proto.Operation{},
	}

	return mp, nil
}
