package p2p

import (
	"encoding/json"

	"github.com/0xsequence/bundler/proto"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (n *Host) HandleMessageType(messageType proto.MessageType, handler func(from peer.ID, message []byte)) {
	n.handlers[messageType] = handler
}

func (n *Host) setupPubsub() error {
	logger := n.logger

	psOptions := []pubsub.Option{
		pubsub.WithMessageSignaturePolicy(pubsub.StrictSign),
		pubsub.WithMaxMessageSize(1 << 20), // 1MB

		// TODO: only use pubsubtracer in debug mode
		pubsub.WithEventTracer(&PubSubTracer{logger: logger}),
	}

	ps, err := pubsub.NewGossipSub(n.ctx, n.host, psOptions...)
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

func (n *Host) pubsubEventHandler() error {
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

			// NOTE: StrictSign message policy ensures that signatures
			// are validated.

			// TODO: consider using pubsubpb with protobuf for message data

			// fmt.Println("From:", msg.GetFrom().String())
			// fmt.Println("ReceivedFrom:", msg.ReceivedFrom.String())
			// fmt.Println("Key:", hexutil.Encode(msg.Key))

			// address, err := PeerIDToETHAddress(msg.GetFrom())
			// if err != nil {
			// 	panic(err)
			// }
			// fmt.Println("ETH ADDRESS OF PEER", address.String())

			// Filter out messages from self
			if msg.GetFrom() == n.host.ID() {
				continue
			}

			var message proto.Message
			err = json.Unmarshal(msg.Data, &message)
			if err != nil {
				n.logger.Info("failed to unmarshal pubsub message", "err", err)
				continue
			}

			n.logger.Info("received pubsub message", "from", msg.GetFrom().String(), "type", message.Type)

			if message.Type != nil {
				handler := n.handlers[*message.Type]
				if handler != nil {
					// TODO: Can't we just not use json.Marshal and directly use msg.Data?
					data, err := json.Marshal(message.Message)
					if err != nil {
						n.logger.Error("unable to marshal message", "err", err)
						continue
					}

					handler(msg.GetFrom(), data)
				} else {
					n.logger.Info("no handler found for message type", "type", *message.Type)
				}
			}
		}
	}()

	return nil
}
