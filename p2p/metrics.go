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

	pubsubReceivedErrors prometheus.Counter
	pubsubFilteredSelf   prometheus.Counter
	pubsubReceivedBytes  *prometheus.HistogramVec
	pubsubHandledTime    *prometheus.HistogramVec

	pubsubAddedPeers     prometheus.Counter
	pubsubRemovedPeers   prometheus.Counter
	pubsubThrottledPeers prometheus.Counter

	pubsubTopicsActive         *prometheus.GaugeVec
	pubsubTopicsGraft          *prometheus.CounterVec
	pubsubTopicsPrune          *prometheus.CounterVec
	pubsubMessageValidate      *prometheus.CounterVec
	pubsubMessageDeliver       *prometheus.CounterVec
	pubsubMessageReject        *prometheus.CounterVec
	pubsubMessageDuplicate     *prometheus.CounterVec
	pubsubMessageUndeliverable *prometheus.CounterVec

	pubsubRPCSubRecv prometheus.Counter
	pubsubRPCPubRecv *prometheus.CounterVec
	pubsubRPCRecv    *prometheus.CounterVec

	pubsubRPCSubSent prometheus.Counter
	pubsubRPCPubSent *prometheus.CounterVec
	pubsubRPCSent    *prometheus.CounterVec

	pubsubRPCSubDrop prometheus.Counter
	pubsubRPCPubDrop *prometheus.CounterVec
	pubsubRPCDrop    *prometheus.CounterVec
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
	}, []string{"topic"})

	pubsubReceivedErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_pubsub_received_errors",
		Help: "Number of pubsub received errors",
	})

	pubsubFilteredSelf := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_pubsub_filtered_self",
		Help: "Number of pubsub messages filtered from self",
	})

	pubsubHandledTime := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "p2p_pubsub_handled_time",
		Help:    "Time taken to handle pubsub messages",
		Buckets: prometheus.ExponentialBuckets(1e-6, 2, 15),
	}, []string{"topic", "result"})

	pubsubReceivedBytes := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "p2p_pubsub_received_bytes",
		Help:    "Number of bytes received in pubsub",
		Buckets: prometheus.ExponentialBuckets(1, 2, 26),
	}, []string{"topic", "result"})

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

	// Gossip metrics
	pubsubAddedPeers := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_pubsub_added_peers_count",
		Help: "Number of peers added to the pubsub system",
	})

	pubsubRemovedPeers := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_pubsub_removed_peers_count",
		Help: "Number of peers removed from the pubsub system",
	})

	pubsubThrottledPeers := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_pubsub_throttled_peers_count",
		Help: "Number of peers throttled in the pubsub system",
	})

	pubsubTopicsActive := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "p2p_pubsub_topics_active",
		Help: "Current number of active topics in the pubsub system",
	}, []string{"topic"})

	pubsubTopicsGraft := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_pubsub_topics_graft",
		Help: "Number of times topics have been grafted",
	}, []string{"topic"})

	pubsubTopicsPrune := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_pubsub_topics_prune",
		Help: "Number of times topics have been pruned",
	}, []string{"topic"})

	pubsubMessageValidate := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_pubsub_message_validate",
		Help: "Number of pubsub messages validated",
	}, []string{"topic"})

	pubsubMessageDeliver := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_pubsub_message_deliver",
		Help: "Number of pubsub messages delivered",
	}, []string{"topic"})

	pubsubMessageReject := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_pubsub_message_reject",
		Help: "Number of pubsub messages rejected",
	}, []string{"topic"})

	pubsubMessageDuplicate := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_pubsub_message_duplicate",
		Help: "Number of duplicate pubsub messages received",
	}, []string{"topic"})

	pubsubMessageUndeliverable := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_pubsub_message_undeliverable",
		Help: "Number of pubsub messages deemed undeliverable",
	}, []string{"topic"})

	pubsubRPCSubRecv := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_pubsub_rpc_sub_recv",
		Help: "Number of subscription RPCs received",
	})

	pubsubRPCPubRecv := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_pubsub_rpc_pub_recv",
		Help: "Number of publish RPCs received",
	}, []string{"topic"})

	pubsubRPCRecv := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_pubsub_rpc_recv",
		Help: "Total number of RPCs received",
	}, []string{"type"})

	pubsubRPCSubSent := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_pubsub_rpc_sub_sent",
		Help: "Number of subscription RPCs sent",
	})

	pubsubRPCPubSent := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_pubsub_rpc_pub_sent",
		Help: "Number of publish RPCs sent",
	}, []string{"topic"})

	pubsubRPCSent := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_pubsub_rpc_sent",
		Help: "Total number of RPCs sent",
	}, []string{"type"})

	pubsubRPCSubDrop := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_pubsub_rpc_sub_drop",
		Help: "Number of subscription RPCs dropped",
	})

	pubsubRPCPubDrop := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_pubsub_rpc_pub_drop",
		Help: "Number of publish RPCs dropped",
	}, []string{"topic"})

	pubsubRPCDrop := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_pubsub_rpc_drop",
		Help: "Total number of RPCs dropped",
	}, []string{"type"})

	if reg != nil {
		reg.MustRegister(
			bootnodesConnected,
			bootnodesFailed,
			bootnodesRetries,
			broadcastErrors,
			broadcastSentBytes,
			pubsubReceivedErrors,
			pubsubFilteredSelf,
			pubsubHandledTime,
			pubsubReceivedBytes,
			foundPeers,
			foundSelfAsPeer,
			foundPeersFailedConnect,
			foundPeersConnected,
			pubsubAddedPeers,
			pubsubRemovedPeers,
			pubsubThrottledPeers,
			pubsubTopicsActive,
			pubsubTopicsGraft,
			pubsubTopicsPrune,
			pubsubMessageValidate,
			pubsubMessageDeliver,
			pubsubMessageReject,
			pubsubMessageDuplicate,
			pubsubMessageUndeliverable,
			pubsubRPCSubRecv,
			pubsubRPCPubRecv,
			pubsubRPCRecv,
			pubsubRPCSubSent,
			pubsubRPCPubSent,
			pubsubRPCSent,
			pubsubRPCSubDrop,
			pubsubRPCPubDrop,
			pubsubRPCDrop,
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

		pubsubReceivedErrors: pubsubReceivedErrors,
		pubsubFilteredSelf:   pubsubFilteredSelf,
		pubsubReceivedBytes:  pubsubReceivedBytes,
		pubsubHandledTime:    pubsubHandledTime,

		pubsubAddedPeers:     pubsubAddedPeers,
		pubsubRemovedPeers:   pubsubRemovedPeers,
		pubsubThrottledPeers: pubsubThrottledPeers,

		pubsubTopicsActive:         pubsubTopicsActive,
		pubsubTopicsGraft:          pubsubTopicsGraft,
		pubsubTopicsPrune:          pubsubTopicsPrune,
		pubsubMessageValidate:      pubsubMessageValidate,
		pubsubMessageDeliver:       pubsubMessageDeliver,
		pubsubMessageReject:        pubsubMessageReject,
		pubsubMessageDuplicate:     pubsubMessageDuplicate,
		pubsubMessageUndeliverable: pubsubMessageUndeliverable,

		pubsubRPCSubRecv: pubsubRPCSubRecv,
		pubsubRPCPubRecv: pubsubRPCPubRecv,
		pubsubRPCRecv:    pubsubRPCRecv,

		pubsubRPCSubSent: pubsubRPCSubSent,
		pubsubRPCPubSent: pubsubRPCPubSent,
		pubsubRPCSent:    pubsubRPCSent,

		pubsubRPCSubDrop: pubsubRPCSubDrop,
		pubsubRPCPubDrop: pubsubRPCPubDrop,
		pubsubRPCDrop:    pubsubRPCDrop,
	}
}
