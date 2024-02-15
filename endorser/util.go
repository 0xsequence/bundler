package endorser

import "github.com/0xsequence/ethkit/go-ethereum/accounts/abi"

func (r *EndorserResult) Encode() ([]byte, error) {
	return args().Pack(r.Readiness, r.BlockDependency, r.Dependencies)
}

func args() abi.Arguments {
	if args_ == nil {
		args_ = abi.Arguments{
			{
				Name: "readiness",
				Type: newType("bool", "bool", nil),
			},
			{
				Name: "blockDependency",
				Type: newType("tuple", "struct Endorser.BlockDependency", []abi.ArgumentMarshaling{
					{
						Name:         "maxNumber",
						Type:         "uint256",
						InternalType: "uint256",
					},
					{
						Name:         "maxTimestamp",
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
