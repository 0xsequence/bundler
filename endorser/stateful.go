package endorser

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/common/hexutil"
)

type DependencyState struct {
	Balance *big.Int   `json:"balance,omitempty"`
	Code    []byte     `json:"code,omitempty"`
	Nonce   *uint64    `json:"nonce,omitempty"`
	Slots   [][32]byte `json:"slots,omitempty"`
}

func (d *Dependency) HasChanged(x, y *DependencyState) (bool, error) {
	if err := d.Validate(x); err != nil {
		return false, fmt.Errorf("x is not a valid state for dependency on %v: %w", d.Addr, err)
	}
	if err := d.Validate(y); err != nil {
		return false, fmt.Errorf("y is not a valid state for dependency on %v: %w", d.Addr, err)
	}

	if d.Balance {
		if x.Balance.Cmp(y.Balance) != 0 {
			return true, nil
		}
	}

	if d.Code {
		if !bytes.Equal(x.Code, y.Code) {
			return true, nil
		}
	}

	if d.Nonce {
		if *x.Nonce != *y.Nonce {
			return true, nil
		}
	}

	return false, nil
}

func (d *Dependency) Validate(state *DependencyState) error {
	if (state.Balance != nil) != d.Balance {
		return fmt.Errorf("balance existence does not match dependency")
	}

	if (state.Code != nil) != d.Code {
		return fmt.Errorf("code existence does not match dependency")
	}

	if (state.Nonce != nil) != d.Nonce {
		return fmt.Errorf("nonce existence does not match dependency")
	}

	if len(state.Slots) != len(d.Slots) {
		return fmt.Errorf("number of slots does not match dependency")
	}

	return nil
}

type EndorserResult struct {
	Readiness        bool             `json:"readiness"`
	GlobalDependency GlobalDependency `json:"global_dependency"`
	Dependencies     []Dependency     `json:"dependencies"`
}

type EndorserResultState struct {
	Dependencies []DependencyState `json:"dependencies"`
}

func (r *EndorserResult) State(ctx context.Context, provider *ethrpc.Provider) (*EndorserResultState, error) {
	state := EndorserResultState{}

	for _, dependency := range r.Dependencies {
		state_ := DependencyState{}

		if dependency.Balance {
			var err error
			state_.Balance, err = provider.BalanceAt(ctx, dependency.Addr, nil)
			if err != nil {
				return nil, fmt.Errorf("unable to read balance for %v: %w", dependency.Addr, err)
			}
		}

		if dependency.Code {
			code, err := provider.CodeAt(ctx, dependency.Addr, nil)
			if err != nil {
				return nil, fmt.Errorf("unable to read code for %v: %w", dependency.Addr, err)
			}
			if code == nil {
				code = []byte{}
			}
			state_.Code = code
		}

		if dependency.Nonce {
			nonce, err := provider.NonceAt(ctx, dependency.Addr, nil)
			if err != nil {
				return nil, fmt.Errorf("unable to read nonce for %v: %w", dependency.Addr, err)
			}
			state_.Nonce = &nonce
		}

		state_.Slots = make([][32]byte, 0, len(dependency.Slots))
		for _, slot := range dependency.Slots {
			value, err := provider.StorageAt(ctx, dependency.Addr, slot, nil)
			if err != nil {
				return nil, fmt.Errorf("unable to read storage for %v at %v: %w", dependency.Addr, hexutil.Encode(slot[:]), err)
			}
			state_.Slots = append(state_.Slots, [32]byte(value))
		}

		state.Dependencies = append(state.Dependencies, state_)
	}

	return &state, nil
}

func (r *EndorserResult) HasChanged(x, y *EndorserResultState) (bool, error) {
	if err := r.Validate(x); err != nil {
		return false, fmt.Errorf("x is not a valid state for endorser result: %w", err)
	}
	if err := r.Validate(y); err != nil {
		return false, fmt.Errorf("y is not a valid state for endorser result: %w", err)
	}

	for i, dependency := range r.Dependencies {
		hasChanged, err := dependency.HasChanged(&x.Dependencies[i], &y.Dependencies[i])
		if err != nil {
			return false, err
		}

		if hasChanged {
			return true, nil
		}
	}

	return false, nil
}

func (r *EndorserResult) Validate(state *EndorserResultState) error {
	if len(state.Dependencies) != len(r.Dependencies) {
		return fmt.Errorf("number of dependencies does not match endorser result")
	}

	for i, dependency := range r.Dependencies {
		if err := dependency.Validate(&state.Dependencies[i]); err != nil {
			return fmt.Errorf("dependency state %v does not match dependency: %w", i, err)
		}
	}

	return nil
}

func (r *EndorserResult) CheckConstraints(ctx context.Context, provider *ethrpc.Provider) (bool, error) {
	for _, dependency := range r.Dependencies {
		for _, constraint := range dependency.Constraints {
			slot := constraint.Slot
			value, err := provider.StorageAt(ctx, dependency.Addr, slot, nil)
			if err != nil {
				return false, fmt.Errorf("unable to read storage for %v at %v: %w", dependency.Addr, hexutil.Encode(slot[:]), err)
			}

			bnMin := new(big.Int).SetBytes(constraint.MinValue[:])
			bnMax := new(big.Int).SetBytes(constraint.MaxValue[:])
			bnValue := new(big.Int).SetBytes(value[:])

			if bnValue.Cmp(bnMin) < 0 || bnValue.Cmp(bnMax) > 0 {
				return false, nil
			}
		}
	}

	return true, nil
}
