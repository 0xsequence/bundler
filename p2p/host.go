package p2p

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/ethkit/ethwallet"
	"github.com/0xsequence/ethkit/go-ethereum/common/hexutil"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	p2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
	"github.com/prometheus/client_golang/prometheus"
)

type Host struct {
	lock sync.Mutex

	cfg     *config.P2PHostConfig
	logger  *slog.Logger
	metrics *metrics
	host    host.Host
	pubsub  *pubsub.PubSub
	topics  map[string]*pubsub.Topic

	peerPrivKey crypto.PrivKey

	chainID *big.Int

	running int32
}

var _ Interface = &Host{}

type Identity struct {
	ID peer.ID

	privKey crypto.PrivKey
}

func NewIdentity(mnmonic string) (*Identity, error) {
	wallet, err := ethwallet.NewWalletFromMnemonic(mnmonic)
	if err != nil {
		return nil, err
	}

	// Use private key at HD node account index 0 as the peer private key.
	peerPrivKeyBytes, err := hexutil.Decode(wallet.PrivateKeyHex())
	if err != nil {
		return nil, err
	}

	// Generate a deterministic private key from the given bytes
	// but use Ed25519, as libp2p does not support secp256k1
	// (it does still work, but with secp256k1 things are unstable)
	peerPrivKey, _, err := crypto.GenerateEd25519Key(bytes.NewReader(peerPrivKeyBytes))
	if err != nil {
		return nil, err
	}

	id, err := peer.IDFromPrivateKey(peerPrivKey)
	if err != nil {
		return nil, err
	}

	return &Identity{
		ID:      id,
		privKey: peerPrivKey,
	}, nil
}

func NewHost(cfg *config.P2PHostConfig, logger *slog.Logger, metrics prometheus.Registerer, identity *Identity, chainID *big.Int) (*Host, error) {
	logger = logger.With("hostId", identity.ID.String())

	connmgr, err := connmgr.NewConnManager(
		300, // Lowwater
		400, // HighWater,
		connmgr.WithGracePeriod(time.Minute),
	)
	if err != nil {
		return nil, err
	}

	var registerOpts libp2p.Option
	if metrics != nil {
		registerOpts = libp2p.PrometheusRegisterer(metrics)
	}

	h, err := libp2p.New(
		// Use the keypair we generated
		libp2p.Identity(identity.privKey),

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
		// libp2p.ListenAddrs(listenAddr),

		// support TLS connections
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		// support noise connections
		libp2p.Security(noise.ID, noise.New),

		// support any other default transports (TCP)
		// libp2p.DefaultTransports,
		libp2p.ChainOptions(
			libp2p.NoTransports,
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
		libp2p.EnableNATService(),

		// ..
		libp2p.DisableRelay(),

		libp2p.EnableHolePunching(),

		// Metrics
		registerOpts,

		// TODO: review all libp2p options and defaults
	)

	if err != nil {
		return nil, err
	}

	nd := &Host{
		cfg:         cfg,
		logger:      logger,
		metrics:     createMetrics(metrics),
		host:        h,
		peerPrivKey: identity.privKey,
		chainID:     chainID,

		topics: make(map[string]*pubsub.Topic),
	}

	return nd, nil
}

func (n *Host) Run(ctx context.Context) error {
	if n.IsRunning() {
		return fmt.Errorf("node: already running")
	}
	atomic.StoreInt32(&n.running, 1)
	defer atomic.StoreInt32(&n.running, 0)

	hostAddrs := getHostAddresses(n.host)
	for _, hostAddr := range hostAddrs {
		n.logger.Info("-> node: listening libp2p", "op", "run", "addr", hostAddr)
	}

	bootPeers, err := AddrInfoFromP2pAddrs(n.cfg.BootNodeAddrs)
	if err != nil {
		n.logger.Error("error while parsing libp2p boot node addrs", "err", err)
		return err
	}

	err = n.bootstrap(ctx, bootPeers)
	if err != nil {
		return err
	}

	priorityPeers, err := AddrInfoFromP2pAddrs(n.cfg.PriorityNodeAddrs)
	if err != nil {
		n.logger.Error("error while parsing libp2p priority node addrs", "err", err)
		return err
	}

	for _, peerInfo := range priorityPeers {
		n.logger.Info("protecting priority peer", "peerId", peerInfo.ID.String())
		n.host.ConnManager().Protect(peerInfo.ID, "priority")
	}

	err = n.setupPubsub(ctx, n.chainID)
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

func (n *Host) bootstrap(ctx context.Context, bootPeers []peer.AddrInfo) error {
	logger := n.logger

	// run a DHT router in server mode
	kdht, err := dht.New(ctx, n.host, dht.Mode(dht.ModeServer))
	if err != nil {
		return err
	}

	connected := make(chan struct{})

	var wg sync.WaitGroup
	for _, pinfo := range bootPeers {
		// n.host.Peerstore().AddAddrs(pinfo.ID, pinfo.Addrs, peerstore.PermanentAddrTTL)
		wg.Add(1)
		go func(pinfo peer.AddrInfo) {
			defer wg.Done()
			err := n.host.Connect(ctx, pinfo)
			if err != nil {
				logger.Warn(fmt.Sprintf("Failed to connect to peer %s", pinfo.ID), "err", err)
				n.metrics.bootnodesFailed.Inc()
				return
			}
			logger.Info(fmt.Sprintf("Connected to peer %s", pinfo.ID), slog.String("peer", pinfo.ID.String()))
			n.metrics.bootnodesConnected.Inc()
			connected <- struct{}{}
		}(pinfo)
	}

	go func() {
		wg.Wait()
		close(connected)
	}()

	i := 0
	for range connected {
		i++
	}
	if nPeers := len(bootPeers); i < nPeers/2 {
		logger.Warn(fmt.Sprintf("only connected to %d bootstrap peers out of %d", i, nPeers))
	}

	err = kdht.Bootstrap(ctx)
	if err != nil {
		logger.Error("kdht bootstrap error", "err", err)
		return err
	}

	// routingDiscovery := discrouting.NewRoutingDiscovery(kdht)
	// discutil.Advertise(context.Background(), routingDiscovery, DiscoveryNameSpace)

	discoveryNameSpaceCid, err := NamespaceToCid(Namespace.For(n.chainID))
	if err != nil {
		return err
	}

	// broadcast our existence at the given namespace every 10 seconds.
	go func() {
		for {
			err := kdht.Provide(context.Background(), discoveryNameSpaceCid, true)
			if err != nil {
				logger.Error("error while providing discovery namespace", "err", err)
			}
			time.Sleep(10 * time.Second)
		}
	}()

	n.host.Network().Notify(&p2pnetwork.NotifyBundle{
		ConnectedF: func(net p2pnetwork.Network, c p2pnetwork.Conn) {
			tagged := n.host.ConnManager().GetTagInfo(c.RemotePeer())
			if tagged == nil {
				return
			}
			if tagged.Tags["discovered"] == 0 {
				return
			}
			logger.Info("notify, connected to namespaced peer", "peerId", c.RemotePeer())
		},
		DisconnectedF: func(net p2pnetwork.Network, c p2pnetwork.Conn) {
			tagged := n.host.ConnManager().GetTagInfo(c.RemotePeer())
			if tagged == nil {
				return
			}
			if tagged.Tags["discovered"] == 0 {
				return
			}
			logger.Info("notify, disconnected from namespaced peer", "peerId", c.RemotePeer())
			n.host.ConnManager().UntagPeer(c.RemotePeer(), "discovered")
		},
	})

	// start discovery process in the background.
	go func() {
		failedPeers := map[peer.ID]time.Time{}

		for {
			// search for peers for just 15 seconds and then try again.
			func() {
				tctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()

				peerCh := kdht.FindProvidersAsync(tctx, discoveryNameSpaceCid, 0)
				for peerInfo := range peerCh {
					if peerInfo.ID == n.host.ID() {
						// do not dial ourselves
						continue
					}

					n.metrics.foundPeers.Inc()

					if n.host.Network().Connectedness(peerInfo.ID) == p2pnetwork.Connected {
						// already connected
						continue
					}

					// Skip peers with no addresses, as they are gone
					// NOTE: we comment this out as failedPeers check below is enough
					// if len(peerInfo.Addrs) == 0 {
					// 	continue
					// }

					// Skip if recently failed peer connection
					if t, ok := failedPeers[peerInfo.ID]; ok {
						if time.Since(t) < 1*time.Minute {
							continue
						} else {
							delete(failedPeers, peerInfo.ID)
						}
					}

					// Connect to the peer
					if err := n.host.Connect(ctx, peerInfo); err != nil {
						n.metrics.foundPeersFailedConnect.Inc()
						n.logger.Warn(fmt.Sprintf("failed to connect with namespaced peer %s", peerInfo.String()), "err", err)
						failedPeers[peerInfo.ID] = time.Now()
						continue
					}

					// tag the peer so that we can offer it higher priority among peers
					n.metrics.foundPeersConnected.Inc()
					n.logger.Info("connected with namespaced peer", "peerId", peerInfo.String())
					n.host.ConnManager().TagPeer(peerInfo.ID, "discovered", 500)
				}
			}()

			// slight delay, and continue searching
			time.Sleep(3 * time.Second)
		}
	}()

	return nil
}

// func (n *Host) attemptBootConnect(ctx context.Context, peerInfo peer.AddrInfo) chan error {
// 	res := make(chan error, 1)

// 	go func(ctx context.Context, peerInfo peer.AddrInfo) {
// 		defer close(res)

// 		for i := 0.0; i <= 25; i += 1.0 {
// 			if ctx.Err() != nil {
// 				res <- fmt.Errorf("context cancelled during boot peer connection attempt")
// 				return
// 			}

// 			if err := n.host.Connect(ctx, peerInfo); err != nil {
// 				// Add a random jitter to avoid synchronized reconnection attempts
// 				jitter := rand.Float64() * i
// 				retryIn := time.Duration(math.Pow(2, i)+jitter) * time.Second
// 				n.metrics.bootnodesRetries.Inc()
// 				n.logger.Error(fmt.Sprintf("error while connecting with boot peer %s", peerInfo.String()), "err", err, "retryIn", retryIn)
// 				time.Sleep(retryIn + time.Duration(float64(retryIn)*rand.Float64()))
// 				continue
// 			}

// 			n.metrics.bootnodesConnected.Inc()
// 			n.logger.Info("connected with the given bootstrap peer", "peerId", peerInfo.String())
// 			res <- nil
// 			return
// 		}

// 		n.metrics.bootnodesFailed.Inc()
// 		res <- fmt.Errorf("failed to connect with boot peer %s", peerInfo.String())
// 	}(ctx, peerInfo)

// 	return res
// }

func (n *Host) HostID() peer.ID {
	if n.host == nil {
		return ""
	} else {
		return n.host.ID()
	}
}

func (n *Host) Address() (string, error) {
	return n.host.ID().String(), nil
}

func (n *Host) Sign(data []byte) ([]byte, error) {
	return n.peerPrivKey.Sign(data)
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
		if tag != nil && tag.Tags["priority"] > 0 {
			priorityPeers = append(priorityPeers, p)
		}
	}
	return priorityPeers
}
