package p2p

import (
	"github.com/0xsequence/bundler/proto"
	"github.com/libp2p/go-libp2p/core/peer"
)

type MsgHandler func(from peer.ID, message []byte)

type Interface interface {
	Broadcast(payload proto.Message) error
	HandleMessageType(messageType proto.MessageType, handler MsgHandler)
}
