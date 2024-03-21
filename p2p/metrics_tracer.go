package p2p

import (
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/prometheus/client_golang/prometheus"
)

var _ = pubsub.RawTracer(metricsTracer{})

// Initializes the values for the pubsub rpc action.
type action int

const (
	recv action = iota
	send
	drop
)

// This tracer is used to implement metrics collection for messages received
// and broadcasted through gossipsub.
type metricsTracer struct {
	metrics *metrics
}

func newMetricsTracer(metrics *metrics) *metricsTracer {
	return &metricsTracer{metrics: metrics}
}

// AddPeer .
func (g metricsTracer) AddPeer(p peer.ID, proto protocol.ID) {
	g.metrics.pubsubAddedPeers.Inc()
}

// RemovePeer .
func (g metricsTracer) RemovePeer(p peer.ID) {
	g.metrics.pubsubRemovedPeers.Inc()
}

// Join .
func (g metricsTracer) Join(topic string) {
	g.metrics.pubsubTopicsActive.WithLabelValues(topic).Set(1)
}

// Leave .
func (g metricsTracer) Leave(topic string) {
	g.metrics.pubsubTopicsActive.WithLabelValues(topic).Set(0)
}

// Graft .
func (g metricsTracer) Graft(p peer.ID, topic string) {
	g.metrics.pubsubTopicsGraft.WithLabelValues(topic).Inc()
}

// Prune .
func (g metricsTracer) Prune(p peer.ID, topic string) {
	g.metrics.pubsubTopicsPrune.WithLabelValues(topic).Inc()
}

// ValidateMessage .
func (g metricsTracer) ValidateMessage(msg *pubsub.Message) {
	g.metrics.pubsubMessageValidate.WithLabelValues(*msg.Topic).Inc()
}

// DeliverMessage .
func (g metricsTracer) DeliverMessage(msg *pubsub.Message) {
	g.metrics.pubsubMessageDeliver.WithLabelValues(*msg.Topic).Inc()
}

// RejectMessage .
func (g metricsTracer) RejectMessage(msg *pubsub.Message, reason string) {
	g.metrics.pubsubMessageReject.WithLabelValues(*msg.Topic, reason).Inc()
}

// DuplicateMessage .
func (g metricsTracer) DuplicateMessage(msg *pubsub.Message) {
	g.metrics.pubsubMessageDuplicate.WithLabelValues(*msg.Topic).Inc()
}

// UndeliverableMessage .
func (g metricsTracer) UndeliverableMessage(msg *pubsub.Message) {
	g.metrics.pubsubMessageUndeliverable.WithLabelValues(*msg.Topic).Inc()
}

// ThrottlePeer .
func (g metricsTracer) ThrottlePeer(p peer.ID) {
	g.metrics.pubsubThrottledPeers.Inc()
}

// RecvRPC .
func (g metricsTracer) RecvRPC(rpc *pubsub.RPC) {
	g.setMetricFromRPC(recv, g.metrics.pubsubRPCSubRecv, g.metrics.pubsubRPCPubRecv, g.metrics.pubsubRPCRecv, rpc)
}

// SendRPC .
func (g metricsTracer) SendRPC(rpc *pubsub.RPC, p peer.ID) {
	g.setMetricFromRPC(send, g.metrics.pubsubRPCSubSent, g.metrics.pubsubRPCPubSent, g.metrics.pubsubRPCSent, rpc)
}

// DropRPC .
func (g metricsTracer) DropRPC(rpc *pubsub.RPC, p peer.ID) {
	g.setMetricFromRPC(drop, g.metrics.pubsubRPCSubDrop, g.metrics.pubsubRPCPubDrop, g.metrics.pubsubRPCDrop, rpc)
}

func (g metricsTracer) setMetricFromRPC(act action, subCtr prometheus.Counter, pubCtr, ctrlCtr *prometheus.CounterVec, rpc *pubsub.RPC) {
	subCtr.Add(float64(len(rpc.Subscriptions)))
	if rpc.Control != nil {
		ctrlCtr.WithLabelValues("graft").Add(float64(len(rpc.Control.Graft)))
		ctrlCtr.WithLabelValues("prune").Add(float64(len(rpc.Control.Prune)))
		ctrlCtr.WithLabelValues("ihave").Add(float64(len(rpc.Control.Ihave)))
		ctrlCtr.WithLabelValues("iwant").Add(float64(len(rpc.Control.Iwant)))
	}
	for _, msg := range rpc.Publish {
		// For incoming messages from pubsub, we do not record metrics for them as these values
		// could be junk.
		if act == recv {
			continue
		}
		pubCtr.WithLabelValues(*msg.Topic).Inc()
	}
}
