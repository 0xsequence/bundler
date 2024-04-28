package admin

import (
	"context"
	"fmt"
	"sort"

	"github.com/0xsequence/bundler/ipfs"
	"github.com/0xsequence/bundler/lib/mempool"
	"github.com/0xsequence/bundler/lib/registry"
	"github.com/0xsequence/bundler/lib/types"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/go-chi/httplog/v2"
)

type Admin struct {
	logger *httplog.Logger

	Ipfs     ipfs.Interface
	Mempool  mempool.Interface
	Registry registry.Interface
}

func NewAdmin(logger *httplog.Logger, ipfs ipfs.Interface, mempool mempool.Interface, registry registry.Interface) *Admin {
	return &Admin{
		logger:   logger,
		Ipfs:     ipfs,
		Mempool:  mempool,
		Registry: registry,
	}
}

func (a Admin) BanEndorser(ctx context.Context, endorser string, duration int) error {
	if !common.IsHexAddress(endorser) {
		return fmt.Errorf("invalid endorser address")
	}

	a.Registry.BanEndorser(common.HexToAddress(endorser), registry.PermanentBan)
	return nil
}

func (a Admin) BannedEndorsers(ctx context.Context) ([]string, error) {
	allEndorsers := a.Registry.KnownEndorsers()
	bannedEndorsers := make([]string, 0, len(allEndorsers))

	for _, endorser := range allEndorsers {
		if endorser.Status == registry.PermanentBanned {
			bannedEndorsers = append(bannedEndorsers, endorser.Address.Hex())
		}
	}

	return bannedEndorsers, nil
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
