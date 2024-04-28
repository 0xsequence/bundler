package rpc

import (
	"context"

	"github.com/0xsequence/bundler/lib/types"
	"github.com/0xsequence/bundler/p2p"
	"github.com/0xsequence/bundler/proto"
)

func (s *RPC) SendOperation(ctx context.Context, pop *proto.Operation) (string, error) {
	op, err := types.NewOperationFromProto(pop)
	if err != nil {
		return "", err
	}

	// Always PIN these operations to IPFS
	// as they are being sent by the user, and
	// it is useful for debugging
	go op.ReportToIPFS(s.ipfs)

	err = s.mempool.AddOperation(ctx, op, true)
	if err != nil {
		return "", err
	}

	// If the operation is fine, broadcast it to the network
	s.Host.Broadcast(ctx, p2p.OperationTopic, op.ToProtoPure())

	return op.Hash(), nil
}

func (s RPC) Mempool(ctx context.Context) (*proto.MempoolView, error) {
	return s.mempool.Inspect(), nil
}

func (s RPC) Operations(ctx context.Context) (*proto.Operations, error) {
	return s.archive.Operations(ctx), nil
}

func (s *RPC) FeeAsks(ctx context.Context) (*proto.FeeAsks, error) {
	return s.collector.FeeAsks()
}
