package endorser

import "github.com/0xsequence/ethkit/go-ethereum/accounts/abi"

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
