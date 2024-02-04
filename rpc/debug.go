package rpc

import "context"

type Debug struct {
	RPC *RPC
}

func (d *Debug) Broadcast(ctx context.Context, message interface{}) (bool, error) {
	err := d.RPC.Host.Broadcast(message)
	if err != nil {
		return false, err
	}
	return true, nil
}
