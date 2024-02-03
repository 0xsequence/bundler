package rpc

import "context"

func (r *RPC) Broadcast(ctx context.Context, message interface{}) (bool, error) {
	err := r.Host.Broadcast(message)
	if err != nil {
		return false, err
	}
	return true, nil
}
