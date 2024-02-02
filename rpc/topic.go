package rpc

import "context"

func (r *RPC) Broadcast(ctx context.Context, data string) (bool, error) {
	err := r.Node.Broadcast([]byte(data))
	if err != nil {
		return false, err
	}
	return true, nil
}
