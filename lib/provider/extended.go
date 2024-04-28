package provider

import (
	"bytes"
	"context"
	"fmt"
	"sync/atomic"

	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
)

type extendedMetrics struct {
	overrideCalls *prometheus.CounterVec

	supportsDebug    prometheus.Gauge
	supportsOverride prometheus.Gauge
}

type Extended struct {
	*ethrpc.Provider

	metrics *extendedMetrics

	supportsDebug    *atomic.Int32 // 0 = unknown, 1 = supported, 2 = not supported
	supportsOverride *atomic.Int32 // 0 = unknown, 1 = supported, 2 = not supported
}

func createMetrics() *extendedMetrics {
	return &extendedMetrics{
		supportsDebug: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "supports_debug",
			Help: "Whether the provider supports the debug RPC method",
		}),
		supportsOverride: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "supports_override",
			Help: "Whether the provider supports the override RPC method",
		}),
		overrideCalls: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "override_calls",
			Help: "Number of override calls made",
		}, []string{"result"}),
	}
}

func (e *extendedMetrics) register(reg prometheus.Registerer) {
	reg.MustRegister(
		e.supportsDebug,
		e.supportsOverride,
		e.overrideCalls,
	)
}

func NewExtendedAuto(provider *ethrpc.Provider) *Extended {
	// TODO Do a call to the provider to check if it supports debug and override
	return &Extended{
		Provider: provider,

		metrics: createMetrics(),

		supportsDebug:    &atomic.Int32{},
		supportsOverride: &atomic.Int32{},
	}
}

func (p *Extended) SetRegisterer(reg prometheus.Registerer) {
	p.metrics.register(reg)
}

func NewExtended(provider *ethrpc.Provider, supportsDebug, supportsOverride bool) *Extended {
	supportsDebugInt := atomic.Int32{}
	if supportsDebug {
		supportsDebugInt.Store(1)
	} else {
		supportsDebugInt.Store(2)
	}

	supportsOverrideInt := atomic.Int32{}
	if supportsOverride {
		supportsOverrideInt.Store(1)
	} else {
		supportsOverrideInt.Store(2)
	}

	metrics := createMetrics()
	metrics.supportsDebug.Set(float64(supportsDebugInt.Load()))
	metrics.supportsOverride.Set(float64(supportsOverrideInt.Load()))

	return &Extended{
		Provider:         provider,
		metrics:          metrics,
		supportsDebug:    &supportsDebugInt,
		supportsOverride: &supportsOverrideInt,
	}
}

func (p *Extended) SupportsDebug() bool {
	return p.supportsDebug.Load() == 1
}

func (p *Extended) SupportsOverride() bool {
	return p.supportsOverride.Load() == 1
}

type Call struct {
	To   common.Address `json:"to"`
	Data string         `json:"data"`
}

type Override struct {
	Code      *string                     `json:"code"`
	StateDiff map[common.Hash]common.Hash `json:"stateDiff"`
}

type OverrideArgs map[common.Address]*Override

func (p *Extended) CallWithOverride(ctx context.Context, call *Call, overrides OverrideArgs) ([]byte, error) {
	if p.supportsOverride.Load() == 2 {
		p.metrics.overrideCalls.WithLabelValues("unsupported").Inc()
		return nil, fmt.Errorf("provider does not support overrides")
	}

	rpcCall := ethrpc.NewCallBuilder[string]("eth_call", nil, call, nil, overrides)
	var res string
	_, err := p.Do(ctx, rpcCall.Into(&res))
	if err != nil {
		// TODO: Did we get an unsupported error?
		// if so, we should move from 0 to 2
		p.metrics.overrideCalls.WithLabelValues("error").Inc()
		return nil, fmt.Errorf("eth_call failed: %w", err)
	}

	p.metrics.overrideCalls.WithLabelValues("success").Inc()
	return common.FromHex(res), nil
}

func (a OverrideArgs) Merge(b OverrideArgs) error {
	for k, v := range b {
		if _, ok := a[k]; ok {
			err := a[k].Merge(v)
			if err != nil {
				return err
			}
		}

		a[k] = v
	}

	return nil
}

func (a *Override) Merge(b *Override) error {
	if b.Code != nil {
		if a.Code != nil && (*a.Code) != (*b.Code) {
			return fmt.Errorf("cannot merge overrides with conflicting code")
		}

		a.Code = b.Code
	}

	for k, v := range b.StateDiff {
		if _, ok := a.StateDiff[k]; ok {
			if !bytes.Equal(a.StateDiff[k].Bytes(), v.Bytes()) {
				return fmt.Errorf("cannot merge overrides with conflicting state diff, key: %s", k)
			}
		}

		a.StateDiff[k] = v
	}

	return nil
}
