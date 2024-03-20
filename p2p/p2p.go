package p2p

import "math/big"

const (
	DiscoveryNamespace = "erc5189-bundler"

	PubsubTopicPrefix = "erc5189-bundler-op-mempool-"
)

func PubsubTopic(chainID *big.Int) string {
	return PubsubTopicPrefix + "-" + chainID.String()
}
