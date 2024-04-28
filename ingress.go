package bundler

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/lib/collector"
	"github.com/0xsequence/bundler/lib/types"
	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/p2p"
	"github.com/0xsequence/bundler/proto"
	"github.com/go-chi/httplog/v2"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/prometheus/client_golang/prometheus"
)

type ingressMetrics struct {
	pendingOps prometheus.Gauge

	processTime prometheus.Histogram

	acceptedOps prometheus.Counter
	droppedOps  *prometheus.CounterVec

	dropReasonUnmarshal    prometheus.Labels
	dropReasonFromProto    prometheus.Labels
	dropReasonLowFee       prometheus.Labels
	dropReasonErrorPayment prometheus.Labels
	dropReasonMempool      prometheus.Labels
}

type Ingress struct {
	logger  *httplog.Logger
	metrics *ingressMetrics

	Host      p2p.Interface
	Mempool   mempool.Interface
	Collector collector.Interface
}

func createIngressMetrics(reg prometheus.Registerer) *ingressMetrics {
	acceptedOps := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ingress_accepted_ops",
		Help: "Number of operations accepted",
	})

	pendingOps := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ingress_pending_ops",
		Help: "Number of operations pending",
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
		reg.MustRegister(
			acceptedOps,
			pendingOps,
			droppedOps,
			processTime,
		)
	}

	return &ingressMetrics{
		acceptedOps: acceptedOps,
		pendingOps:  pendingOps,
		droppedOps:  droppedOps,
		processTime: processTime,

		dropReasonUnmarshal:    prometheus.Labels{"reason": "unmarshal"},
		dropReasonFromProto:    prometheus.Labels{"reason": "from_proto"},
		dropReasonLowFee:       prometheus.Labels{"reason": "low_fee"},
		dropReasonErrorPayment: prometheus.Labels{"reason": "error_payment"},
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
		logger:  logger,
		metrics: createIngressMetrics(metrics),

		Host:      host,
		Mempool:   mempool,
		Collector: collector,
	}
}

func (i *Ingress) registerHandler(ctx context.Context) {
	i.Host.HandleTopic(ctx, p2p.OperationTopic, func(ctx context.Context, p peer.ID, data []byte) pubsub.ValidationResult {
		i.metrics.pendingOps.Inc()
		i.logger.Info("ingress: received operation")

		start := time.Now()

		defer func() {
			i.metrics.pendingOps.Dec()
			i.metrics.processTime.Observe(time.Since(start).Seconds())
		}()

		// Try to parse the operation
		var protoOp proto.Operation
		err := json.Unmarshal(data, &protoOp)
		if err != nil {
			i.metrics.droppedOps.With(i.metrics.dropReasonUnmarshal).Inc()
			i.logger.Warn("invalid operation message - parse proto", "err", err)
			return pubsub.ValidationReject
		}

		// Try to convert the operation
		op, err := types.NewOperationFromProto(&protoOp)
		if err != nil {
			i.metrics.droppedOps.With(i.metrics.dropReasonFromProto).Inc()
			i.logger.Warn("invalid operation message - parse operation", "err", err)
			return pubsub.ValidationReject
		}

		// Pass it trough the collector, since
		// it can quickly reject it if it doesn't
		// pay enough fees. Don't reject it, only ignore it
		// since it may be a valid operation.
		if err := i.Collector.ValidatePayment(op); err != nil {
			if errors.Is(err, collector.InsufficientFeeError) {
				i.logger.Info("%v", err)
				i.metrics.droppedOps.With(i.metrics.dropReasonLowFee).Inc()
				return pubsub.ValidationIgnore
			} else {
				i.metrics.droppedOps.With(i.metrics.dropReasonErrorPayment).Inc()
				return pubsub.ValidationIgnore
			}
		}

		// If the mempool reject it, only ignore it
		// as it may be a valid operation, our mempool
		// may be full or the operation may have been invalidated
		// in flight.
		err = i.Mempool.AddOperation(ctx, op, false)
		if err != nil {
			i.metrics.droppedOps.With(i.metrics.dropReasonMempool).Inc()
			i.logger.Info("ingress: rejected by the mempool", "error", err, "op", op.Hash())
			return pubsub.ValidationIgnore
		}

		i.metrics.acceptedOps.Inc()
		return pubsub.ValidationAccept
	})

	i.logger.Info("ingress: handler registered")
}

func (i *Ingress) Run(ctx context.Context) error {
	i.registerHandler(ctx)

	return nil
}
