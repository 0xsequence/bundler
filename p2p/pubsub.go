package p2p

import (
	"encoding/json"

	"github.com/davecgh/go-spew/spew"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

func (n *Node) setupPubsub() error {
	logger := n.logger

	// TODO: only use pubsubtracer in debug mode
	ps, err := pubsub.NewGossipSub(n.ctx, n.host, pubsub.WithEventTracer(&PubSubTracer{logger: logger}))
	if err != nil {
		logger.Error("unable to create gossip pub sub", "err", err)
		return err
	}
	topic, err := ps.Join(PubsubTopic)
	if err != nil {
		logger.Error("while creating pub sub topic", "err", err)
		return err
	}

	n.pubsub = ps
	n.topic = topic

	// start the pubsub event handler
	err = n.pubsubEventHandler()
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) pubsubEventHandler() error {
	n.logger.Info("starting pubsub event handler")

	sub, err := n.topic.Subscribe()
	if err != nil {
		n.logger.Error("while creating pubsub subscriber", "err", err)
		return err
	}

	// start receiving gossip message from other peers.
	go func() {
		for {
			select {
			case <-n.ctx.Done():
				sub.Cancel()
				return
			default:
			}

			msg, err := sub.Next(n.ctx)
			if err != nil {
				n.logger.Error("while receving pubsub message", "err", err)
				continue
			}

			// TODO: in the future, lets check the peerId of msg.ReceivedFrom,
			// and potentially the signature too, so we can ignore any malicious peers.

			// TODO: consider using pubsubpb with protobuf for message data

			var data interface{}
			err = json.Unmarshal(msg.Data, &data)
			if err != nil {
				n.logger.Info("failed to unmarshal pubsub message", "err", err)
				continue
			}

			// TODO: parse msg.Data, etc. etc.. check the kind of event, etc.
			spew.Dump(data)

		}
	}()

	return nil
}
