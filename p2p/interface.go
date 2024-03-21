package p2p

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

type MsgHandler func(ctx context.Context, p peer.ID, msg *pubsub.Message) pubsub.ValidationResult

type Interface interface {
	BroadcastData(ctx context.Context, topic PubsubTopic, payload []byte) error
	Broadcast(ctx context.Context, topic PubsubTopic, payload interface{}) error
	HandleTopic(ctx context.Context, topic PubsubTopic, handler MsgHandler) error
	Address() (string, error)
	Sign(data []byte) ([]byte, error)
}
