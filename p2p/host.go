package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/ethkit/ethwallet"
	"github.com/0xsequence/ethkit/go-ethereum/common/hexutil"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
)

type Host struct {
	cfg      *config.Config
	logger   *slog.Logger
	host     host.Host
	pubsub   *pubsub.PubSub
	topic    *pubsub.Topic
	handlers map[proto.MessageType]func(message any)

	ctx     context.Context
	ctxStop context.CancelFunc
	running int32
	// mu      sync.RWMutex
}

func NewHost(cfg *config.Config, logger *slog.Logger, wallet *ethwallet.Wallet) (*Host, error) {

	// Use private key at HD node account index 0 as the peer private key.
	peerPrivKeyBytes, err := hexutil.Decode(wallet.PrivateKeyHex())
	if err != nil {
		return nil, err
	}

	peerPrivKey, err := crypto.UnmarshalSecp256k1PrivateKey(peerPrivKeyBytes)
	if err != nil {
		return nil, err
	}

	id, err := peer.IDFromPrivateKey(peerPrivKey)
	if err != nil {
		return nil, err
	}
	logger = logger.With("hostId", id.String())

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
		libp2p.Identity(peerPrivKey),

		// if we want network to be separated, etc.
		// libp2p.PrivateNetwork(),

		// Multiple listen addresses
		//
		// Addr hosts result in, for example:
		// /ip4/127.0.0.1/tcp/5000/p2p/16Uiu2HAmKygtVwc8pYhcHPbAJidkLtNce4Mge6eFu3fZpB7Vu3ri
		// /ip4/127.0.0.1/udp/5000/quic-v1/p2p/16Uiu2HAmKygtVwc8pYhcHPbAJidkLtNce4Mge6eFu3fZpB7Vu3ri
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", cfg.P2PPort),         // TCP transport
			fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", cfg.P2PPort), // QUIC transport
		),

		// support TLS connections
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		// support noise connections
		libp2p.Security(noise.ID, noise.New),

		// support any other default transports (TCP)
		// libp2p.DefaultTransports,
		libp2p.ChainOptions(
			libp2p.Transport(tcp.NewTCPTransport),
			libp2p.Transport(quic.NewTransport),
		),

		// Let's prevent our peer from having too many
		// connections by attaching a connection manager.
		libp2p.ConnectionManager(connmgr),

		// Attempt to open ports using uPNP for NATed hosts.
		libp2p.NATPortMap(),

		// If you want to help other peers to figure out if they are behind
		// NATs, you can launch the server-side of AutoNAT too (AutoRelay
		// already runs the client)
		//
		// This service is highly rate-limited and should not cause any
		// performance issues.
		// libp2p.EnableNATService(),

		// ..
		libp2p.DisableRelay(),

		// TODO: review all libp2p options and defaults
	)

	if err != nil {
		return nil, err
	}

	nd := &Host{
		cfg:      cfg,
		logger:   logger,
		host:     h,
		handlers: map[proto.MessageType]func(message any){},
	}

	return nd, nil
}

func (n *Host) Run(ctx context.Context) error {
	if n.IsRunning() {
		return fmt.Errorf("node: already running")
	}
	atomic.StoreInt32(&n.running, 1)
	defer atomic.StoreInt32(&n.running, 0)

	n.ctx, n.ctxStop = context.WithCancel(ctx)

	hostAddr := getHostAddress(n.host)
	n.logger.Info(fmt.Sprintf("-> node: listening on %s", hostAddr), "op", "run", "addr", hostAddr)

	bootPeers, err := AddrInfoFromP2pAddrs(n.cfg.BootNodeAddrs)
	if err != nil {
		n.logger.Error("error while parsing libp2p boot node addrs", "err", err)
		return err
	}

	err = n.bootstrap(bootPeers)
	if err != nil {
		return err
	}

	err = n.setupPubsub()
	if err != nil {
		return err
	}

	return nil
}

func (n *Host) Stop(timeoutCtx context.Context) {
	if !n.IsRunning() || n.IsStopping() {
		return
	}
	atomic.StoreInt32(&n.running, 2)
	defer atomic.StoreInt32(&n.running, 0)

	n.logger.Info("-> node: stopping..", "op", "stop")

	// .. any cleanup/stop here
	// ..

	n.logger.Info("-> rpc: stopped.", "op", "stop")
}

func (s *Host) IsRunning() bool {
	return atomic.LoadInt32(&s.running) == 1
}

func (s *Host) IsStopping() bool {
	return atomic.LoadInt32(&s.running) == 2
}

func (n *Host) bootstrap(bootPeers []peer.AddrInfo) error {
	// run a DHT router in server mode
	kdht, err := dht.New(n.ctx, n.host, dht.Mode(dht.ModeServer))
	if err != nil {
		return err
	}

	if err = kdht.Bootstrap(n.ctx); err != nil {
		return err
	}

	// connect with bootstrap peers
	for _, bootPeer := range bootPeers {
		if err := n.host.Connect(n.ctx, bootPeer); err != nil {
			n.logger.Error(fmt.Sprintf("error while connecting with boot peer %s", bootPeer.String()), "err", err)
			continue
		}
		n.logger.Info("connected with the given bootstrap peer", "peerId", bootPeer.String())
	}

	// advertise our existence at the given namespace.
	routingDiscovery := drouting.NewRoutingDiscovery(kdht)
	dutil.Advertise(n.ctx, routingDiscovery, DiscoveryNamespace)

	// start discovery process in the background.
	discoveryNameSpaceCid, err := NamespaceToCid(DiscoveryNamespace)
	if err != nil {
		return err
	}

	go func() {
		peerCh := kdht.FindProvidersAsync(n.ctx, discoveryNameSpaceCid, 0)
		for peerInfo := range peerCh {
			if peerInfo.ID == n.host.ID() {
				// do not dial ourselves
				continue
			}

			if err := n.host.Connect(n.ctx, peerInfo); err != nil {
				n.logger.Error(fmt.Sprintf("failed to connect with namespaced peer %s", peerInfo.String()), "err", err)
				continue
			}

			// tag the peer so that we can offer it higher priority among peers
			n.logger.Info("connected with namespaced peer", "peerId", peerInfo.String())
			n.host.ConnManager().TagPeer(peerInfo.ID, DiscoveryNamespace, 500)
		}
	}()

	return nil
}

func (n *Host) HostID() peer.ID {
	if n.host == nil {
		return ""
	} else {
		return n.host.ID()
	}
}

func (n *Host) Addrs() []multiaddr.Multiaddr {
	if n.host == nil {
		return []multiaddr.Multiaddr{}
	} else {
		return n.host.Network().ListenAddresses()
	}
}

func (n *Host) Peers() []peer.ID {
	if n.host == nil {
		return []peer.ID{}
	} else {
		return n.host.Network().Peers()
	}
}

func (n *Host) PriorityPeers() []peer.ID {
	if n.host == nil {
		return []peer.ID{}
	}

	priorityPeers := []peer.ID{}
	for _, p := range n.host.Network().Peers() {
		tag := n.host.ConnManager().GetTagInfo(p)
		if tag != nil && tag.Tags[DiscoveryNamespace] > 0 {
			priorityPeers = append(priorityPeers, p)
		}
	}
	return priorityPeers
}

func (n *Host) Broadcast(payload proto.Message) error {
	if n.topic == nil {
		return fmt.Errorf("pubsub topic not initialized")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return n.topic.Publish(n.ctx, data)
}
