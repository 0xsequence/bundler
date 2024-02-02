package rpc

import "context"

func (r *RPC) Broadcast(ctx context.Context, data interface{}) (bool, error) {
	err := r.Node.Broadcast(data)
	if err != nil {
		return false, err
	}
	return true, nil
}
