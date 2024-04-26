package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (n *Host) setupPubsub(ctx context.Context, chainID *big.Int) error {
	logger := n.logger

	psOptions := []pubsub.Option{
		pubsub.WithMessageSignaturePolicy(pubsub.StrictSign),
		pubsub.WithMaxMessageSize(1 << 20), // 1MB

		// TODO: only use pubsubtracer in debug mode
		pubsub.WithEventTracer(&PubSubTracer{logger: logger}),
		pubsub.WithRawTracer(newMetricsTracer(n.metrics)),
	}

	ps, err := pubsub.NewGossipSub(ctx, n.host, psOptions...)
	if err != nil {
		logger.Error("unable to create gossip pub sub", "err", err)
		return err
	}

	logger.Info("-> setup pubsub")

	n.chainID = chainID
	n.pubsub = ps
	return nil
}

func (n *Host) waitPubsub(ctx context.Context, topic string) error {
	// wait for pubsub to be ready
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	logged := false

	for {
		select {
		case <-ctx.Done():
			n.logger.Info("-> closing pubsub wait", "topic", topic)
			return fmt.Errorf("context done")
		case <-ticker.C:
			if n.pubsub != nil {
				return nil
			} else {
				if !logged {
					n.logger.Info("-> waiting for pubsub to be ready", "topic", topic)
					logged = true
				}
			}
		}
	}
}

func (n *Host) Broadcast(ctx context.Context, topic PubsubTopic, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return n.BroadcastData(ctx, topic, dataBytes)
}

func (n *Host) BroadcastData(ctx context.Context, topic PubsubTopic, data []byte) error {
	subtopic := topic.For(n.chainID)
	reg, ok := n.topics[subtopic]
	if !ok {
		n.metrics.broadcastErrors.Inc()
		return fmt.Errorf("topic %s not found", subtopic)
	}

	err := reg.Publish(ctx, data)
	if err != nil {
		n.metrics.broadcastErrors.Inc()
		n.logger.Error("while broadcasting pubsub message", "topic", topic, "err", err)
	}

	n.metrics.broadcastSentBytes.WithLabelValues(subtopic).Observe(float64(len(data)))

	return err
}

func (n *Host) HandleTopic(ctx context.Context, topic PubsubTopic, handler MsgHandler) error {
	n.lock.Lock()
	defer n.lock.Unlock()

	subtopic := topic.For(n.chainID)

	err := n.waitPubsub(ctx, subtopic)
	if err != nil {
		return err
	}

	reg, err := n.pubsub.Join(subtopic)
	if err != nil {
		n.logger.Error("while creating pub sub topic", "err", err)
		return err
	}

	sub, err := reg.Subscribe()
	if err != nil {
		n.logger.Error("while subscribing to pub sub topic", "err", err)
		return err
	}

	n.topics[subtopic] = reg

	// It seems hacky to use RegisterTopicValidator to handle messages
	// but validating the messages is expensive, this way we can validate and use them
	// in one go.
	//
	// This could alternatively be handled by removing the verification logic from within
	// the mempool, this way "validator" would be handled by the endorser and the mempool
	// will blindly accept operations.
	sid := n.host.ID()
	err = n.pubsub.RegisterTopicValidator(subtopic, func(ctx context.Context, p peer.ID, msg *pubsub.Message) pubsub.ValidationResult {
		// Do not validate our own messages
		// or else everything takes double the time
		if p == sid {
			n.metrics.pubsubFilteredSelf.Inc()
			return pubsub.ValidationAccept
		}

		start := time.Now()
		res := handler(ctx, p, msg.Data)

		n.metrics.pubsubReceivedBytes.WithLabelValues(subtopic, fmt.Sprint(res)).Observe(float64(len(msg.Data)))
		n.metrics.pubsubHandledTime.WithLabelValues(subtopic, fmt.Sprint(res)).Observe(time.Since(start).Seconds())

		return res
	})

	if err != nil {
		n.logger.Error("while registering pubsub validator", "err", err)
		return err
	}

	// Consume all messages from the subscription
	// they are handled by the registered validator
	go func() {
		for {
			select {
			case <-ctx.Done():
				n.logger.Info("-> closing pubsub subscription", "topic", subtopic)
				sub.Cancel()
				n.logger.Info("-> closing pubsub topic", "topic", subtopic)
				reg.Close()
				n.logger.Info("-> closed pubsub", "topic", subtopic)
				return
			default:
			}

			_, err := sub.Next(ctx)
			if err != nil {
				n.metrics.pubsubReceivedErrors.Inc()
				n.logger.Error("while receiving pubsub message", "err", err)
				continue
			}
		}
	}()

	return nil
}
