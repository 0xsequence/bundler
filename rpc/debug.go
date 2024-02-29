package rpc

import (
	"context"

	"github.com/0xsequence/bundler/proto"
)

type Debug struct {
	RPC *RPC
}

func (d *Debug) Broadcast(ctx context.Context, message interface{}) (bool, error) {
	err := d.RPC.Host.Broadcast(proto.Message{
		Type:    proto.MessageType_DEBUG,
		Message: message,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}
