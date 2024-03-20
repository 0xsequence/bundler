package p2p

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	bootnodesConnected prometheus.Counter
	bootnodesFailed    prometheus.Counter
	bootnodesRetries   prometheus.Counter

	foundPeers              prometheus.Counter
	foundSelfAsPeer         prometheus.Counter
	foundPeersFailedConnect prometheus.Counter
	foundPeersConnected     prometheus.Counter

	broadcastErrors    prometheus.Counter
	broadcastSentBytes *prometheus.HistogramVec

	pubsubReceivedErrors  prometheus.Counter
	pubsubFilteredSelf    prometheus.Counter
	pubsubFailedUnmarshal prometheus.Counter
	pubsubUnhandledMsg    prometheus.Counter
	pubsubReceivedBytes   *prometheus.HistogramVec
	pubsubHandledTime     *prometheus.HistogramVec
}

func createMetrics(reg prometheus.Registerer) *metrics {
	bootnodesConnected := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_bootnodes_connected",
		Help: "Number of bootnodes connected",
	})

	bootnodesFailed := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_bootnodes_failed",
		Help: "Number of bootnodes failed to connect",
	})

	bootnodesRetries := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_bootnodes_retries",
		Help: "Number of bootnodes connection retries",
	})

	broadcastErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_broadcast_errors",
		Help: "Number of broadcast errors",
	})

	broadcastSentBytes := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "p2p_broadcast_sent_bytes",
		Help:    "Number of bytes sent in broadcast",
		Buckets: prometheus.ExponentialBuckets(1, 2, 26),
	}, []string{"type"})

	pubsubReceivedErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_pubsub_received_errors",
		Help: "Number of pubsub received errors",
	})

	pubsubFilteredSelf := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_pubsub_filtered_self",
		Help: "Number of pubsub messages filtered from self",
	})

	pubsubFailedUnmarshal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_pubsub_failed_unmarshal",
		Help: "Number of pubsub messages failed to unmarshal",
	})

	pubsubUnhandledMsg := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_pubsub_unhandled_msg",
		Help: "Number of pubsub messages unhandled",
	})

	pubsubHandledTime := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "p2p_pubsub_handled_time",
		Help:    "Time taken to handle pubsub messages",
		Buckets: prometheus.ExponentialBuckets(1e-6, 2, 15),
	}, []string{"type"})

	pubsubReceivedBytes := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "p2p_pubsub_received_bytes",
		Help:    "Number of bytes received in pubsub",
		Buckets: prometheus.ExponentialBuckets(1, 2, 26),
	}, []string{"type"})

	foundPeers := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_found_peers",
		Help: "Number of peers found",
	})

	foundSelfAsPeer := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_found_self_as_peer",
		Help: "Number of times found self as peer",
	})

	foundPeersFailedConnect := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_found_peers_failed_connect",
		Help: "Number of peers found but failed to connect",
	})

	foundPeersConnected := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_found_peers_connected",
		Help: "Number of peers found and connected",
	})

	if reg != nil {
		reg.MustRegister(
			bootnodesConnected,
			bootnodesFailed,
			bootnodesRetries,
			broadcastErrors,
			broadcastSentBytes,
			pubsubReceivedErrors,
			pubsubFilteredSelf,
			pubsubFailedUnmarshal,
			pubsubUnhandledMsg,
			pubsubHandledTime,
			pubsubReceivedBytes,
			foundPeers,
			foundSelfAsPeer,
			foundPeersFailedConnect,
			foundPeersConnected,
		)
	}

	return &metrics{
		bootnodesConnected: bootnodesConnected,
		bootnodesFailed:    bootnodesFailed,
		bootnodesRetries:   bootnodesRetries,

		broadcastErrors:    broadcastErrors,
		broadcastSentBytes: broadcastSentBytes,

		foundPeers:              foundPeers,
		foundSelfAsPeer:         foundSelfAsPeer,
		foundPeersFailedConnect: foundPeersFailedConnect,
		foundPeersConnected:     foundPeersConnected,

		pubsubReceivedErrors:  pubsubReceivedErrors,
		pubsubFilteredSelf:    pubsubFilteredSelf,
		pubsubFailedUnmarshal: pubsubFailedUnmarshal,
		pubsubUnhandledMsg:    pubsubUnhandledMsg,
		pubsubReceivedBytes:   pubsubReceivedBytes,
		pubsubHandledTime:     pubsubHandledTime,
	}
}
