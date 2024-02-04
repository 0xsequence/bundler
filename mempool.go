package bundler

import "github.com/0xsequence/bundler/proto"

type Mempool struct {
	Operations []proto.Operation `json:"operations"`
}
