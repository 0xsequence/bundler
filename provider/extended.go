package provider

import (
	"bytes"
	"context"
	"fmt"
	"sync/atomic"

	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/common"
)

type Extended struct {
	*ethrpc.Provider

	supportsDebug    *atomic.Int32 // 0 = unknown, 1 = supported, 2 = not supported
	supportsOverride *atomic.Int32 // 0 = unknown, 1 = supported, 2 = not supported
}

func NewExtendedAuto(provider *ethrpc.Provider) *Extended {
	// TODO Do a call to the provider to check if it supports debug and override

	return &Extended{
		Provider:         provider,
		supportsDebug:    &atomic.Int32{},
		supportsOverride: &atomic.Int32{},
	}
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

	return &Extended{
		Provider:         provider,
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
		return nil, fmt.Errorf("provider does not support overrides")
	}

	rpcCall := ethrpc.NewCallBuilder[string]("eth_call", nil, call, nil, overrides)
	var res string
	_, err := p.Do(ctx, rpcCall.Into(&res))
	if err != nil {
		// TODO: Did we get an unsupported error?
		// if so, we should move from 0 to 2
		return nil, fmt.Errorf("eth_call failed: %w", err)
	}

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
