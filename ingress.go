package bundler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/0xsequence/bundler/collector"
	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/p2p"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/types"
	"github.com/go-chi/httplog/v2"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	inputOps    prometheus.Counter
	acceptedOps prometheus.Counter

	pendingOps prometheus.Gauge

	processTime prometheus.Histogram

	droppedOps             *prometheus.CounterVec
	dropReasonKnown        prometheus.Labels
	dropReasonLowFee       prometheus.Labels
	dropReasonErrorPayment prometheus.Labels
	dropReasonInTransit    prometheus.Labels
	dropReasonMempool      prometheus.Labels
}

type Ingress struct {
	handlerRegistered bool

	lock      sync.Mutex
	buffer    chan *types.Operation
	intransit map[string]struct{}

	logger  *httplog.Logger
	metrics *metrics

	Host      p2p.Interface
	Mempool   mempool.Interface
	Collector collector.Interface
}

func createMetrics(reg prometheus.Registerer) *metrics {
	inputOps := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ingress_input_count",
		Help: "Number of operations received",
	})

	pendingOps := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ingress_pending_ops",
		Help: "Number of operations pending",
	})

	acceptedOps := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ingress_accepted_count",
		Help: "Number of operations accepted by the mempool",
	})

	droppedOps := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ingress_drop_count",
		Help: "Number of operations dropped",
	}, []string{"reason"})

	processTime := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "ingress_process_time",
		Help:    "Time it takes to process an operation",
		Buckets: prometheus.DefBuckets,
	})

	if reg != nil {
		reg.MustRegister(inputOps)
		reg.MustRegister(droppedOps)
		reg.MustRegister(pendingOps)
		reg.MustRegister(acceptedOps)
		reg.MustRegister(processTime)
	}

	return &metrics{
		inputOps:    inputOps,
		pendingOps:  pendingOps,
		acceptedOps: acceptedOps,
		droppedOps:  droppedOps,
		processTime: processTime,

		dropReasonKnown:        prometheus.Labels{"reason": "known"},
		dropReasonLowFee:       prometheus.Labels{"reason": "low_fee"},
		dropReasonErrorPayment: prometheus.Labels{"reason": "error_payment"},
		dropReasonInTransit:    prometheus.Labels{"reason": "in_transit"},
		dropReasonMempool:      prometheus.Labels{"reason": "mempool"},
	}
}

func NewIngress(
	cfg *config.MempoolConfig,
	logger *httplog.Logger,
	metrics prometheus.Registerer,
	mempool mempool.Interface,
	collector collector.Interface,
	host p2p.Interface,
) *Ingress {
	return &Ingress{
		lock:      sync.Mutex{},
		buffer:    make(chan *types.Operation, cfg.IngressSize),
		intransit: make(map[string]struct{}, cfg.IngressSize),

		logger:  logger,
		metrics: createMetrics(metrics),

		Host:      host,
		Mempool:   mempool,
		Collector: collector,
	}
}

func (i *Ingress) InBuffer() int {
	return len(i.buffer)
}

func (i *Ingress) registerHandler() {
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
			i.logger.Warn("failed to add operation", "err", err, "op", operation.Hash())
		}
	})
}

func (i *Ingress) Add(op *types.Operation) error {
	i.metrics.inputOps.Inc()

	// If on the mempool known list, we should ignore it
	if i.Mempool.IsKnownOp(op) {
		i.metrics.droppedOps.With(i.metrics.dropReasonKnown).Inc()
		return nil
	}

	// Pass it trough the collector, since
	// it can quickly reject it if it doesn't
	// pay enough fees
	if err := i.Collector.ValidatePayment(op); err != nil {
		if errors.Is(err, collector.InsufficientFeeError) {
			i.logger.Info("%v", err)
			i.metrics.droppedOps.With(i.metrics.dropReasonLowFee).Inc()
			return nil
		} else {
			i.metrics.droppedOps.With(i.metrics.dropReasonErrorPayment).Inc()
			return err
		}
	}

	i.lock.Lock()
	defer i.lock.Unlock()

	// If in transit we should ignore it
	if _, ok := i.intransit[op.Hash()]; ok {
		i.metrics.droppedOps.With(i.metrics.dropReasonInTransit).Inc()
		return nil
	}

	i.metrics.pendingOps.Inc()

	select {
	case i.buffer <- op:
		i.intransit[op.Hash()] = struct{}{}
		return nil
	default:
		return fmt.Errorf("ingress: buffer full")
	}
}

func (i *Ingress) Run(ctx context.Context) {
	i.registerHandler()

	for {
		select {
		case op := <-i.buffer:
			i.metrics.pendingOps.Dec()

			start := time.Now()
			err := i.Mempool.AddOperation(ctx, op, false)
			i.metrics.processTime.Observe(time.Since(start).Seconds())

			if err != nil {
				i.metrics.droppedOps.With(i.metrics.dropReasonMempool).Inc()
				i.logger.Warn("ingress: failed to promote operation", "error", err, "op", op.Hash())
			} else {
				i.metrics.acceptedOps.Inc()
			}

			i.lock.Lock()
			delete(i.intransit, op.Hash())
			i.lock.Unlock()
		case <-ctx.Done():
			return
		}
	}
}
