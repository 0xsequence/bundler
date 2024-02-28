package rpc

import (
	"context"

	"github.com/0xsequence/bundler/proto"
)

type Debug struct {
	RPC *RPC
}

func (d *Debug) Broadcast(ctx context.Context, message interface{}) (bool, error) {
	messageType := proto.MessageType_DEBUG
	err := d.RPC.Host.Broadcast(proto.Message{
		Type:    &messageType,
		Message: message,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}
