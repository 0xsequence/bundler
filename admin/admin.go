package admin

import (
	"context"
	"sort"

	"github.com/0xsequence/bundler/ipfs"
	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/types"
	"github.com/go-chi/httplog/v2"
)

type Admin struct {
	logger *httplog.Logger

	Ipfs    ipfs.Interface
	Mempool mempool.Interface
}

func NewAdmin(logger *httplog.Logger, ipfs ipfs.Interface, mempool mempool.Interface) *Admin {
	return &Admin{
		logger:  logger,
		Ipfs:    ipfs,
		Mempool: mempool,
	}
}

func (a Admin) BanEndorser(ctx context.Context, endorser string, duration int) error {
	panic("unimplemented")
}

func (a Admin) BannedEndorsers(ctx context.Context) ([]string, error) {
	panic("unimplemented")
}

func (a Admin) ReserveOperations(ctx context.Context, num int, skip int, strategy *proto.OperationStrategy) ([]*proto.Operation, error) {
	ops := a.Mempool.ReserveOps(ctx, func(to []*mempool.TrackedOperation) []*mempool.TrackedOperation {
		var ops []*mempool.TrackedOperation

		// If strategy is defined, we need to sort the operations based on the strategy
		// we copy the slice to avoid modifying the original slice
		if strategy != nil {
			toCopy := make([]*mempool.TrackedOperation, len(to))
			copy(toCopy, to)

			switch *strategy {
			case proto.OperationStrategy_Greedy:
				a.logger.Warn("admin: reserve operations: greedy strategy is not supported")
			case proto.OperationStrategy_Fresh:
				// Sort by ReadyAt
				sort.Slice(toCopy, func(i, j int) bool {
					return toCopy[i].ReadyAt.Before(toCopy[j].ReadyAt)
				})
			default:
				a.logger.Warn("admin: reserve operations: unknown strategy")
			}

			ops = toCopy
		}

		if len(ops) > skip {
			return []*mempool.TrackedOperation{}
		}

		ops = ops[skip:]

		if len(ops) > num {
			return ops[:num]
		}

		return ops
	})

	protoOps := make([]*proto.Operation, len(ops))
	for i, op := range ops {
		protoOps[i] = op.ToProto()
	}

	return protoOps, nil
}

func (a Admin) DiscardOperations(ctx context.Context, operations []string) error {
	a.Mempool.DiscardOps(ctx, operations)
	return nil
}

func (a Admin) ReleaseOperations(ctx context.Context, operations []string, readyAtChange *proto.ReadyAtChange) error {
	a.Mempool.ReleaseOps(ctx, operations, *readyAtChange)
	return nil
}

func (a Admin) SendOperation(ctx context.Context, pop *proto.Operation, ignorePayment *bool) (string, error) {
	op, err := types.NewOperationFromProto(pop)
	if err != nil {
		return "", err
	}

	// TODO: Handle ignore payment

	// Always PIN these operations to IPFS
	// as they are being sent by the user, and
	// it is useful for debugging
	go op.ReportToIPFS(a.Ipfs)

	err = a.Mempool.AddOperation(ctx, op, true)
	if err != nil {
		return "", err
	}

	return op.Hash(), nil
}

var _ proto.Admin = Admin{}
