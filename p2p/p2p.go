package p2p

// TODO: potentially rename p2p package to `node`
// and merge with `server` package.

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/ethkit/go-ethereum/common/hexutil"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsubpb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/multiformats/go-multiaddr"
	mh "github.com/multiformats/go-multihash"
)

var DiscoveryNameSpace = "sequence-bundler"

var DiscoveryNameSpaceCid cid.Cid

var PubSubTopic = "erc5189-bundler"

func init() {
	var err error
	DiscoveryNameSpaceCid, err = nsToCid(DiscoveryNameSpace)
	if err != nil {
		panic(err)
	}
}

func nsToCid(ns string) (cid.Cid, error) {
	h, err := mh.Sum([]byte(ns), mh.SHA2_256, -1)
	if err != nil {
		return cid.Undef, err
	}
	return cid.NewCidV1(cid.Raw, h), nil
}

type Node struct {
	cfg    *config.Config
	logger *slog.Logger
	host   host.Host
	pubsub *pubsub.PubSub
	topic  *pubsub.Topic
	ctx    context.Context
}

func NewNode(cfg *config.Config, logger *slog.Logger) (*Node, error) {

	// TODO: support for mnemonic + path ?
	privKey, err := hexutil.Decode(cfg.PrivateKey)
	if err != nil {
		return nil, err
	}

	priv, err := crypto.UnmarshalSecp256k1PrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	id, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		return nil, err
	}
	logger = logger.With("peerId", id.String())

	connmgr, err := connmgr.NewConnManager(
		100, // Lowwater
		400, // HighWater,
		connmgr.WithGracePeriod(time.Minute),
	)
	if err != nil {
		return nil, err
	}

	h, err := libp2p.New(
		// Use the keypair we generated
		libp2p.Identity(priv),

		// if we want network to be separated, etc.
		// libp2p.PrivateNetwork(),

		// Multiple listen addresses
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", cfg.P2PPort), // regular tcp connections
			// fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", cfg.P2PPort), // a UDP endpoint for the QUIC transport
		),
		// support TLS connections
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		// support noise connections
		libp2p.Security(noise.ID, noise.New),
		// support any other default transports (TCP)
		// libp2p.DefaultTransports,
		// Let's prevent our peer from having too many
		// connections by attaching a connection manager.
		libp2p.ConnectionManager(connmgr),
		// Attempt to open ports using uPNP for NATed hosts.
		libp2p.NATPortMap(),

		// Let this host use the DHT to find other hosts
		// libp2p.Routing(func(h host.Host) (corerouting.PeerRouting, error) {
		// 	idht, err = dht.New(ctx, h)
		// 	return idht, err
		// }),

		// If you want to help other peers to figure out if they are behind
		// NATs, you can launch the server-side of AutoNAT too (AutoRelay
		// already runs the client)
		//
		// This service is highly rate-limited and should not cause any
		// performance issues.
		libp2p.EnableNATService(),
	)

	if err != nil {
		return nil, err
	}

	nd := &Node{
		cfg:    cfg,
		logger: logger,
		host:   h,
		ctx:    context.Background(), // TODO: setup in Run()..
	}

	return nd, nil
}

func (n *Node) Run(ctx context.Context) error {
	// TODO: .. standard pattern of Run/Stop/IsRuning, etc..

	logger := n.logger

	pb, err := pubsub.NewGossipSub(ctx, n.host, pubsub.WithEventTracer(&PubSubTracer{logger: logger}))
	if err != nil {
		logger.Error("unable to create gossip pub sub", "err", err)
		return err
	}
	topic, err := pb.Join(PubSubTopic)
	if err != nil {
		logger.Error("while creating pub sub topic", "err", err)
		return err
	}

	n.pubsub = pb
	n.topic = topic
	logger.Info("host started running", "addr", getHostAddress(n.host))

	// run a DHT router in server mode.
	kademliaDHT, err := dht.New(ctx, n.host, dht.Mode(dht.ModeServer))
	if err != nil {
		return err
	}

	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		return err
	}

	// connect with bootstrap node.
	for i, bootNode := range n.cfg.BootNodeAddrs {
		peerInfo, err := peer.AddrInfoFromP2pAddr(bootNode)
		if err != nil {
			n.logger.Error("error while parsing p2p addr", "boot_node", n.cfg.BootNodes[i], "err", err)
			continue
		}
		if err := n.host.Connect(ctx, *peerInfo); err != nil {
			n.logger.Error(fmt.Sprintf("error while connecting with boot node %s", peerInfo.String()), "err", err)
		}
		n.logger.Info("connected with the given bootstrap node", "peerId", peerInfo.String())
	}

	// advertise our existence at the given namespace.
	routingDiscovery := routing.NewRoutingDiscovery(kademliaDHT)
	dutil.Advertise(ctx, routingDiscovery, DiscoveryNameSpace)

	// start discovery process in the background.
	go func() {
		peerCh := kademliaDHT.FindProvidersAsync(ctx, DiscoveryNameSpaceCid, 0)
		for peerInfo := range peerCh {
			if peerInfo.ID == n.host.ID() {
				// do not dial ourselves
				continue
			}

			fmt.Println("found namespaced peer!! connecting to peer..", peerInfo) // TODO: use logger
			if err := n.host.Connect(context.Background(), peerInfo); err != nil {
				n.logger.Error(fmt.Sprintf("error while connecting with boot node %s", peerInfo.String()), "err", err)
			} else {
				// ..
				n.host.ConnManager().TagPeer(peerInfo.ID, "keep-namespaced-peer", 100)
			}
			n.logger.Info("connected with discovered node", "peerId", peerInfo.String())
		}
	}()

	// TODO........
	go n.StartEventHandler()

	return nil
}

func (n *Node) Stop() {
	// TODO ... copy pattern..
}

func (n *Node) Bootstrap(peers []peer.AddrInfo) {
	// ..
}

func (n *Node) Broadcast(data []byte) error {
	ctx := context.Background() // TODO: use n.ctx etc.
	return n.topic.Publish(ctx, data)
}

func (n *Node) StartEventHandler() error {
	n.logger.Info("starting event handler")

	sub, err := n.topic.Subscribe()
	if err != nil {
		n.logger.Error("while creating pubsub subscriber", "err", err)
		return err
	}

	// start receiving gossip message from other peers.
	go func() {
		for {
			msg, err := sub.Next(n.ctx)
			if err != nil {
				n.logger.Error("while receving pubsub message", "err", err)
				continue
			}
			fmt.Println("local msg..?", msg.Local, msg.ReceivedFrom.String())
			fmt.Println("msg data:", string(msg.Data))
			// spew.Dump(msg)
		}
	}()

	return nil
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
		t.logger.Info("new peer added", "id", peerID.String())
	default:
		t.logger.Debug("trace", "event", evt)
	}
}

func (n *Node) Peers() []peer.ID {
	if n.host == nil {
		return []peer.ID{}
	} else {
		return n.host.Network().Peers()
	}
}

func (n *Node) Addrs() []multiaddr.Multiaddr {
	if n.host == nil {
		return []multiaddr.Multiaddr{}
	} else {
		return n.host.Network().ListenAddresses()
	}
}

func (n *Node) PeerID() peer.ID {
	if n.host == nil {
		return ""
	} else {
		return n.host.ID()
	}
}
