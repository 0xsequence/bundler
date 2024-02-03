package p2p

import (
	"fmt"
	"log/slog"

	"github.com/0xsequence/ethkit/ethwallet"
	"github.com/0xsequence/ethkit/go-ethereum/accounts"
	"github.com/0xsequence/ethkit/go-ethereum/common/hexutil"
	"github.com/ipfs/go-cid"
	pubsubpb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	mh "github.com/multiformats/go-multihash"
)

func setupWallet(privateKeyHex string, derivationPath string) (*ethwallet.Wallet, []byte, error) {
	privKey, err := hexutil.Decode(privateKeyHex)
	if err != nil {
		return nil, nil, err
	}

	derivPath := accounts.DefaultBaseDerivationPath
	if derivationPath != "" {
		derivPath, err = ethwallet.ParseDerivationPath(derivationPath)
		if err != nil {
			return nil, nil, err
		}
	}

	hdnode, err := ethwallet.NewHDNodeFromEntropy(privKey, &derivPath)
	if err != nil {
		return nil, nil, err
	}
	// fmt.Println("hdnode", hdnode.Address().String(), hdnode.DerivationPath().String())

	// Create ethereum HD wallet used by the txn senders.
	wallet, err := ethwallet.NewWalletFromHDNode(hdnode)
	if err != nil {
		return nil, nil, err
	}

	// Use private key at HD node account index 0 as the peer private key.
	peerPrivKeyBytes, err := hexutil.Decode(wallet.PrivateKeyHex())
	if err != nil {
		return nil, nil, err
	}

	return wallet, peerPrivKeyBytes, nil
}

func nsToCid(ns string) (cid.Cid, error) {
	h, err := mh.Sum([]byte(ns), mh.SHA2_256, -1)
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
