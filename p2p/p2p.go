package p2p

// TODO: potentially rename p2p package to `node`
// and merge with `server` package.

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsubpb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/multiformats/go-multiaddr"
)

var PubSubTopic = "erc5189-bundler"

type Node struct {
	cfg    *config.Config
	logger *slog.Logger
	host   host.Host
	pubsub *pubsub.PubSub
	topic  *pubsub.Topic
	ctx    context.Context
}

func NewNode(cfg *config.Config, logger *slog.Logger) (*Node, error) {

	// TODO/NOTE: will change to use the config's full length
	// private key, or mnemonic + path
	var seedKey io.Reader
	if cfg.SeedKey != "" {
		key := make([]byte, 64) // 64 byte key length
		copy(key, []byte(cfg.SeedKey))
		seedKey = bytes.NewReader(key)
	} else {
		// TODO: not a good source of randomness
		seedKey = rand.Reader
	}

	priv, _, err := crypto.GenerateKeyPairWithReader(
		crypto.Secp256k1,
		-1,
		seedKey,
	)
	if err != nil {
		return nil, err
	}

	connmgr, err := connmgr.NewConnManager(
		100, // Lowwater
		400, // HighWater,
		connmgr.WithGracePeriod(time.Minute),
	)
	if err != nil {
		panic(err)
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

	ctx := context.Background() // TODO..

	nd := &Node{
		logger: logger,
		host:   h,
		cfg:    cfg,
		ctx:    ctx,
	}

	emitter, err := h.EventBus().Emitter(new(pubsubpb.TraceEvent))
	if err != nil {
		return nil, err
	}

	pb, err := pubsub.NewGossipSub(ctx, h, pubsub.WithEventTracer(&PubSubTracer{emitter: emitter}))
	if err != nil {
		logger.Error("unable to create gossip pub sub", "err", err)
		return nil, err
	}
	topic, err := pb.Join(PubSubTopic)
	if err != nil {
		logger.Error("while creating pub sub topic", "err", err)
		return nil, err
	}

	nd.pubsub = pb
	nd.topic = topic
	logger.Info("host started running", "addr", getHostAddress(h))

	return nd, nil
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
	logger  *slog.Logger
	emitter event.Emitter
}

func (t *PubSubTracer) Trace(evt *pubsubpb.TraceEvent) {
	switch *evt.Type {
	case pubsubpb.TraceEvent_ADD_PEER:
		t.logger.Info("new peer added", "id", string(evt.AddPeer.PeerID))
		go func() {
			t.emitter.Emit(*evt)
		}()
	default:
		t.logger.Debug("trace", "event", evt)
	}
}
