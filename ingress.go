package bundler

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/p2p"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/types"
	"github.com/go-chi/httplog/v2"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Ingress struct {
	handlerRegistered bool

	lock      sync.Mutex
	buffer    chan *types.Operation
	intransit map[string]struct{}

	logger *httplog.Logger

	Host    *p2p.Host
	Mempool *Mempool
}

func NewIngress(cfg *config.MempoolConfig, logger *httplog.Logger, m *Mempool, h *p2p.Host) *Ingress {
	return &Ingress{
		lock:      sync.Mutex{},
		buffer:    make(chan *types.Operation, cfg.IngressSize),
		intransit: make(map[string]struct{}, cfg.IngressSize),

		logger: logger,

		Host:    h,
		Mempool: m,
	}
}

func (i *Ingress) RegisterHanler() {
	if i.handlerRegistered {
		return
	}

	i.handlerRegistered = true
	i.Host.HandleMessageType(proto.MessageType_NEW_OPERATION, func(_ peer.ID, message []byte) {
		var protoOperation proto.Operation
		err := json.Unmarshal(message, &protoOperation)
		if err != nil {
			// TODO: Mark peer as bad
			i.logger.Warn("invalid operation message - parse proto", "err", err)
			return
		}

		operation, err := types.NewOperationFromProto(&protoOperation)
		if err != nil {
			// TODO: Mark peer as bad
			i.logger.Warn("invalid operation message - parse operation", "err", err)
			return
		}

		err = i.Add(operation)
		if err != nil {
			i.logger.Warn("failed to add operation", "err", err, "op", operation.Digest())
		}
	})
}

func (i *Ingress) Add(op *types.Operation) error {
	// If on the mempool known list, we should ignore it
	if i.Mempool.IsKnownOp(op) {
		return nil

	}

	i.lock.Lock()
	defer i.lock.Unlock()

	// If in transit we should ignore it
	if _, ok := i.intransit[op.Digest()]; ok {
		return nil
	}

	select {
	case i.buffer <- op:
		i.intransit[op.Digest()] = struct{}{}
		return nil
	default:
		return fmt.Errorf("ingress: buffer full")
	}
}

func (i *Ingress) Run(ctx context.Context) {
	i.RegisterHanler()

	for {
		select {
		case op := <-i.buffer:
			err := i.Mempool.tryPromoteOperation(ctx, op)
			if err != nil {
				i.logger.Warn("ingress: failed to promote operation", "error", err, "op", op.Digest())
			}

			i.lock.Lock()
			delete(i.intransit, op.Digest())
			i.lock.Unlock()
		case <-ctx.Done():
			return
		}
	}
}