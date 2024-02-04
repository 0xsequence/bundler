package p2p

import (
	"fmt"
	"log/slog"

	"github.com/0xsequence/ethkit/go-ethereum/common"
	ethcrypto "github.com/0xsequence/ethkit/go-ethereum/crypto"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/ipfs/go-cid"
	pubsubpb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	mh "github.com/multiformats/go-multihash"
)

func PeerIDToETHAddress(peerID peer.ID) (common.Address, error) {
	pubKey, err := peerID.ExtractPublicKey()
	if err != nil {
		return common.Address{}, err
	}

	raw, _ := pubKey.Raw()
	k, err := secp256k1.ParsePubKey(raw)
	if err != nil {
		return common.Address{}, err
	}

	return ethcrypto.PubkeyToAddress(*k.ToECDSA()), nil
}

func NamespaceToCid(namespace string) (cid.Cid, error) {
	h, err := mh.Sum([]byte(namespace), mh.SHA2_256, -1)
	if err != nil {
		return cid.Undef, err
	}
	return cid.NewCidV1(cid.Raw, h), nil
}

func AddrInfoFromP2pAddrs(multiaddrs []multiaddr.Multiaddr) ([]peer.AddrInfo, error) {
	peerInfos := make([]peer.AddrInfo, 0, len(multiaddrs))
	for _, addr := range multiaddrs {
		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return nil, err
		}
		peerInfos = append(peerInfos, *peerInfo)
	}
	return peerInfos, nil
}

func getHostAddress(ha host.Host) string {
	// Build host multiaddress
	hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", ha.ID().String()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addr := ha.Addrs()[0]
	return addr.Encapsulate(hostAddr).String()
}

type PubSubTracer struct {
	logger *slog.Logger
}

func (t *PubSubTracer) Trace(evt *pubsubpb.TraceEvent) {
	switch *evt.Type {
	case pubsubpb.TraceEvent_ADD_PEER:
		peerID, err := peer.IDFromBytes(evt.AddPeer.PeerID)
		if err != nil {
			panic(err)
		}
		t.logger.Debug("new peer added", "id", peerID.String())
	default:
		t.logger.Debug("trace", "event", evt)
	}
}
