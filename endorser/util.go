package endorser

import (
	"bytes"
	"fmt"

	"github.com/0xsequence/bundler/contracts/gen/solabis/abiendorser"
	"github.com/0xsequence/bundler/contracts/gen/solabis/abivalidator"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi"
)

func (r *EndorserResult) Encode() ([]byte, error) {
	return args().Pack(r.Readiness, r.GlobalDependency, r.Dependencies)
}

func args() abi.Arguments {
	if args_ == nil {
		args_ = abi.Arguments{
			{
				Name: "readiness",
				Type: newType("bool", "bool", nil),
			},
			{
				Name: "globalDependency",
				Type: newType("tuple", "struct Endorser.GlobalDependency", []abi.ArgumentMarshaling{
					{
						Name:         "basefee",
						Type:         "bool",
						InternalType: "bool",
					},
					{
						Name:         "blobbasefee",
						Type:         "bool",
						InternalType: "bool",
					},
					{
						Name:         "chainid",
						Type:         "bool",
						InternalType: "bool",
					},
					{
						Name:         "coinbase",
						Type:         "bool",
						InternalType: "bool",
					},
					{
						Name:         "difficulty",
						Type:         "bool",
						InternalType: "bool",
					},
					{
						Name:         "gasLimit",
						Type:         "bool",
						InternalType: "bool",
					},
					{
						Name:         "number",
						Type:         "bool",
						InternalType: "bool",
					},
					{
						Name:         "timestamp",
						Type:         "bool",
						InternalType: "bool",
					},
					{
						Name:         "txOrigin",
						Type:         "bool",
						InternalType: "bool",
					},
					{
						Name:         "txGasPrice",
						Type:         "bool",
						InternalType: "bool",
					},
					{
						Name:         "maxBlockNumber",
						Type:         "uint256",
						InternalType: "uint256",
					},
					{
						Name:         "maxBlockTimestamp",
						Type:         "uint256",
						InternalType: "uint256",
					},
				}),
			},
			{
				Name: "dependencies",
				Type: newType("tuple[]", "struct Endorser.Dependency[]", []abi.ArgumentMarshaling{
					{
						Name:         "addr",
						Type:         "address",
						InternalType: "address",
					},
					{
						Name:         "balance",
						Type:         "bool",
						InternalType: "bool",
					},
					{
						Name:         "code",
						Type:         "bool",
						InternalType: "bool",
					},
					{
						Name:         "nonce",
						Type:         "bool",
						InternalType: "bool",
					},
					{
						Name:         "allSlots",
						Type:         "bool",
						InternalType: "bool",
					},
					{
						Name:         "slots",
						Type:         "bytes32[]",
						InternalType: "bytes32[]",
					},
					{
						Name:         "constraints",
						Type:         "tuple[]",
						InternalType: "struct Endorser.Constraint[]",
						Components: []abi.ArgumentMarshaling{
							{
								Name:         "slot",
								Type:         "bytes32",
								InternalType: "bytes32",
							},
							{
								Name:         "minValue",
								Type:         "bytes32",
								InternalType: "bytes32",
							},
							{
								Name:         "maxValue",
								Type:         "bytes32",
								InternalType: "bytes32",
							},
						},
					},
				}),
			},
		}
	}

	return args_
}

var args_ abi.Arguments

func newType(t string, internalType string, components []abi.ArgumentMarshaling) abi.Type {
	type_, _ := abi.NewType(t, internalType, components)
	return type_
}

func HasChanged(d *abiendorser.EndorserDependency, x, y *DependencyState) (bool, error) {
	if err := Validate(d, x); err != nil {
		return false, fmt.Errorf("x is not a valid state for dependency on %v: %w", d.Addr, err)
	}
	if err := Validate(d, y); err != nil {
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

func Validate(d *abiendorser.EndorserDependency, state *DependencyState) error {
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

func (r *EndorserResult) HasChanged(x, y *EndorserResultState) (bool, error) {
	if err := r.Validate(x); err != nil {
		return false, fmt.Errorf("x is not a valid state for endorser result: %w", err)
	}
	if err := r.Validate(y); err != nil {
		return false, fmt.Errorf("y is not a valid state for endorser result: %w", err)
	}

	for i, dependency := range r.Dependencies {
		hasChanged, err := HasChanged(&dependency, &x.Dependencies[i], &y.Dependencies[i])
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
		if err := Validate(&dependency, &state.Dependencies[i]); err != nil {
			return fmt.Errorf("dependency state %v does not match dependency: %w", i, err)
		}
	}

	return nil
}

func FromExecutorResult(r *abivalidator.OperationValidatorSimulationResult) *EndorserResult {
	globalDependency := abiendorser.EndorserGlobalDependency{
		Basefee:           r.GlobalDependency.Basefee,
		Blobbasefee:       r.GlobalDependency.Blobbasefee,
		Chainid:           r.GlobalDependency.Chainid,
		Coinbase:          r.GlobalDependency.Coinbase,
		Difficulty:        r.GlobalDependency.Difficulty,
		GasLimit:          r.GlobalDependency.GasLimit,
		Number:            r.GlobalDependency.Number,
		Timestamp:         r.GlobalDependency.Timestamp,
		TxOrigin:          r.GlobalDependency.TxOrigin,
		TxGasPrice:        r.GlobalDependency.TxGasPrice,
		MaxBlockNumber:    r.GlobalDependency.MaxBlockNumber,
		MaxBlockTimestamp: r.GlobalDependency.MaxBlockTimestamp,
	}

	dependencies := make([]abiendorser.EndorserDependency, len(r.Dependencies))
	for i, d := range r.Dependencies {
		dependencies[i] = abiendorser.EndorserDependency{
			Addr:     d.Addr,
			Balance:  d.Balance,
			Code:     d.Code,
			Nonce:    d.Nonce,
			AllSlots: d.AllSlots,
			Slots:    d.Slots,
		}
	}

	return &EndorserResult{
		Readiness:        r.Readiness,
		GlobalDependency: globalDependency,
		Dependencies:     dependencies,
	}
}
