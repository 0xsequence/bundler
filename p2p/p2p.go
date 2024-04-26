package p2p

import "math/big"

type DiscoveryNamespace string

const Namespace = DiscoveryNamespace("ERC5189:pool")

type PubsubTopic string

const OperationTopic = PubsubTopic("ERC5189:pool:op")
const ArchiveTopic = PubsubTopic("ERC5189:pool:archive")

func (p PubsubTopic) For(chainID *big.Int) string {
	return string(p) + ":" + chainID.String()
}

func (d DiscoveryNamespace) For(chainID *big.Int) string {
	return string(d) + ":" + chainID.String()
}
