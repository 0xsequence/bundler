package p2p

import "math/big"

const (
	DiscoveryNamespace = "erc5189-mempool"

	PubsubTopicPrefix = "erc5189-mempool"
)

func PubsubTopic(chainID *big.Int) string {
	return PubsubTopicPrefix + "-" + chainID.String()
}
